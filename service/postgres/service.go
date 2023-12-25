package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gerdooshell/financial/environment"
	"github.com/gerdooshell/financial/lib/cache/lrucache"
	"github.com/gerdooshell/financial/lib/database"
	"github.com/gerdooshell/financial/lib/database/connection_config"
	"github.com/gerdooshell/financial/lib/database/database_engine"
	"github.com/gerdooshell/financial/lib/database/database_hosttag"
	"github.com/gerdooshell/financial/service/postgres/tables"
	bpaPlugin "github.com/gerdooshell/financial/tax/canada/bpa/abstract_plugin"
	cppPlugin "github.com/gerdooshell/financial/tax/canada/canada_pension_plan/abstract_plugin"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	fedPlugin "github.com/gerdooshell/financial/tax/canada/federal/abstract_plugin"
	eiPlugin "github.com/gerdooshell/financial/tax/canada/federal_ei_premium/abstract_plugin"
	prPlugin "github.com/gerdooshell/financial/tax/canada/provincial/abstract_plugin"
	tbPlugin "github.com/gerdooshell/financial/tax/canada/rrsp/abstract_tax_brackets_plugin"
	"os"
	"sync"
)

type BpaCache struct {
	Year   int
	Region canada_region.RegionCode
}
type fedCache struct {
	Year   int
	Salary float32
}
type prCache struct {
	Year   int
	Region canada_region.Province
	Salary float32
}
type cppCache struct{ Year int }
type eiCache struct{ Year int }

type fedBracketCache struct{ Year int }
type prBracketCache struct {
	Year     int
	Province canada_region.Province
}

type pgServiceImpl struct {
	database.ConnectionPool
	bpaCache         lrucache.LRUCache[BpaCache]
	fedCache         lrucache.LRUCache[fedCache]
	prCache          lrucache.LRUCache[prCache]
	cppCache         lrucache.LRUCache[cppCache]
	eiCache          lrucache.LRUCache[eiCache]
	fedBracketsCache lrucache.LRUCache[fedBracketCache]
	prBracketCache   lrucache.LRUCache[prBracketCache]
	bpaMu            sync.Mutex
	fedMu            sync.Mutex
	prMu             sync.Mutex
	cppMu            sync.Mutex
	eiMu             sync.Mutex
	fedBracketsMu    sync.Mutex
	prBracketsMu     sync.Mutex
}

type PGService interface {
	prPlugin.TaxService
	fedPlugin.TaxService
	cppPlugin.CppService
	bpaPlugin.BPAService
	eiPlugin.FedEiService
	tbPlugin.TaxBracketsService
}

var pgCache = lrucache.NewLRUCache[struct {
	Engine  database_engine.Engine
	HostTag database_hosttag.HostTag
}](5)

var mu sync.Mutex

func NewPGService() (PGService, error) {
	mu.Lock()
	defer mu.Unlock()
	env := environment.GetEnvironment()
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	conf, err := connection_config.FromConfigFile(fmt.Sprintf("%v/service/postgres/config.yaml", path), env)
	if err != nil {
		return nil, err
	}
	var srv *pgServiceImpl
	poolGeneric, err := pgCache.Read(conf.GetSignature())
	if err == nil {
		srv = poolGeneric.(*pgServiceImpl)
	} else {
		pool, err := database.NewConnectionPool(conf)
		if err != nil {
			return nil, err
		}
		srv = &pgServiceImpl{ConnectionPool: *pool,
			bpaCache:         lrucache.NewLRUCache[BpaCache](5),
			fedCache:         lrucache.NewLRUCache[fedCache](5),
			prCache:          lrucache.NewLRUCache[prCache](10),
			cppCache:         lrucache.NewLRUCache[cppCache](5),
			eiCache:          lrucache.NewLRUCache[eiCache](5),
			fedBracketsCache: lrucache.NewLRUCache[fedBracketCache](2),
			prBracketCache:   lrucache.NewLRUCache[prBracketCache](15),
		}
		pgCache.Add(conf.GetSignature(), srv)
	}
	return srv, nil
}

func (p *pgServiceImpl) GetBPAEntities(params []*bpaPlugin.BpaPluginInputParameters) (<-chan *bpaPlugin.BpaEntity, error) {
	out := make(chan *bpaPlugin.BpaEntity)
	go func() {
		p.bpaMu.Lock()
		defer p.bpaMu.Unlock()
		defer close(out)
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()

		for _, param := range params {
			var entity *bpaPlugin.BpaEntity
			cacheKey := BpaCache{Year: param.Year, Region: param.RegionCode}
			value, err := p.bpaCache.Read(cacheKey)
			if err == nil {
				entity = value.(*bpaPlugin.BpaEntity)
			} else {
				var row tables.BasicPersonalAmountModel
				transaction.Model(tables.BasicPersonalAmountModel{}).
					Where("year = ?", param.Year).
					Where("region_code = ?", param.RegionCode).Find(&row)
				entity = row.ToBpaEntity()
				p.bpaCache.Add(cacheKey, entity)
			}

			out <- entity
		}
	}()

	return out, nil
}

func (p *pgServiceImpl) GetProvincialTaxEntities(ctx context.Context, params []*prPlugin.TaxCalculatorPluginInputParameters) (<-chan []*prPlugin.TaxEntity, error) {
	out := make(chan []*prPlugin.TaxEntity)
	go func(out chan []*prPlugin.TaxEntity) {
		defer close(out)
		p.prMu.Lock()
		defer p.prMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		for _, param := range params {
			var entities []*prPlugin.TaxEntity
			cacheKey := prCache{Year: param.Year, Region: param.ProvinceCode, Salary: param.Salary}
			if value, err := p.prCache.Read(cacheKey); err == nil {
				entities = value.([]*prPlugin.TaxEntity)
			} else {
				rows := make([]tables.CanadaTaxBracketsModel, 0)
				transaction.Model(tables.CanadaTaxBracketsModel{}).
					Where("year = ?", param.Year).
					Where("region_code = ?", param.ProvinceCode).
					Where("min_salary < ?", param.Salary).Find(&rows)
				entities = make([]*prPlugin.TaxEntity, 0, 10)
				for _, row := range rows {
					entity := row.ToProvincialTaxEntity()
					entities = append(entities, entity)
				}
				p.prCache.Add(cacheKey, entities)
			}
			out <- entities
		}
	}(out)
	return out, nil
}

func (p *pgServiceImpl) GetFederalTaxEntities(params []*fedPlugin.TaxCalculatorPluginInputParameters) (<-chan []*fedPlugin.TaxEntity, error) {
	out := make(chan []*fedPlugin.TaxEntity)
	go func(out chan []*fedPlugin.TaxEntity) {
		defer close(out)
		p.fedMu.Lock()
		defer p.fedMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		var entities []*fedPlugin.TaxEntity
		for _, param := range params {
			cacheKey := fedCache{Year: param.Year, Salary: param.Salary}
			if value, err := p.fedCache.Read(cacheKey); err == nil {
				entities = value.([]*fedPlugin.TaxEntity)
			} else {
				rows := make([]tables.CanadaTaxBracketsModel, 0)
				transaction.Model(tables.CanadaTaxBracketsModel{}).
					Where("year = ?", param.Year).
					Where("region_code = 'ca'").
					Where("min_salary < ?", param.Salary).Order("min_salary ASC").Find(&rows)

				entities = make([]*fedPlugin.TaxEntity, 0, 10)
				for _, row := range rows {
					entity := row.ToFederalTaxEntity()
					entities = append(entities, entity)
				}
				p.fedCache.Add(cacheKey, entities)
			}

			out <- entities
		}
	}(out)
	return out, nil
}

func (p *pgServiceImpl) GetCppEntities(params []*cppPlugin.CppPluginInputParameters) (<-chan *cppPlugin.CppEntity, error) {
	out := make(chan *cppPlugin.CppEntity)
	go func(out chan *cppPlugin.CppEntity) {
		defer close(out)
		p.cppMu.Lock()
		defer p.cppMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		var entity *cppPlugin.CppEntity
		for _, param := range params {
			cacheKey := cppCache{Year: param.Year}
			if value, err := p.cppCache.Read(cacheKey); err == nil {
				entity = value.(*cppPlugin.CppEntity)
			} else {
				row := tables.CanadaPensionPlanModel{}
				transaction.Model(tables.CanadaPensionPlanModel{}).
					Where("Year = ?", param.Year).First(&row)
				entity = row.ToCppEntity()
				p.cppCache.Add(cacheKey, entity)
			}

			out <- entity
		}
	}(out)
	return out, nil
}

func (p *pgServiceImpl) GetEIEntities(params []*eiPlugin.FedEiPluginInputParameters) (<-chan *eiPlugin.FedEiEntity, error) {
	out := make(chan *eiPlugin.FedEiEntity)
	go func() {
		defer close(out)
		p.eiMu.Lock()
		defer p.eiMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		var entity *eiPlugin.FedEiEntity
		for _, param := range params {
			cacheKey := eiCache{Year: param.Year}
			if value, err := p.eiCache.Read(cacheKey); err == nil {
				entity = value.(*eiPlugin.FedEiEntity)
			} else {
				row := tables.FederalEmploymentInsuranceModel{}
				transaction.Model(tables.FederalEmploymentInsuranceModel{}).
					Where("Year = ?", param.Year).First(&row)
				entity = row.ToEiEntity()
				p.eiCache.Add(cacheKey, entity)
			}
			out <- entity
		}
	}()
	return out, nil
}

func (p *pgServiceImpl) GetFedTaxBrackets(params []*tbPlugin.FedTaxBracketsInputParams) (<-chan *tbPlugin.FedTaxBracketOutputParams, error) {
	out := make(chan *tbPlugin.FedTaxBracketOutputParams)
	go func() {
		defer close(out)
		p.fedBracketsMu.Lock()
		defer p.fedBracketsMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		var output *tbPlugin.FedTaxBracketOutputParams
		for _, param := range params {
			cacheKey := fedBracketCache{Year: param.Year}
			if value, err := p.fedBracketsCache.Read(cacheKey); err == nil {
				output = value.(*tbPlugin.FedTaxBracketOutputParams)
			} else {
				rows := make([]tables.CanadaTaxBracketsModel, 0)
				transaction.Model(tables.CanadaTaxBracketsModel{}).
					Where("year = ?", param.Year).
					Where("region_code = ?", param.Region).Order("min_salary ASC").Find(&rows)
				entities := make([]*tbPlugin.TaxBracketEntity, 0, len(rows))
				for _, row := range rows {
					entities = append(entities, row.ToTaxBracketEntity())
				}
				output = &tbPlugin.FedTaxBracketOutputParams{
					FedTaxBracketsInputParams: *param,
					Entities:                  entities,
				}
				p.fedBracketsCache.Add(cacheKey, output)
			}
			out <- output
		}
	}()
	return out, nil
}

func (p *pgServiceImpl) GetPrTaxBrackets(params []*tbPlugin.PrTaxBracketsInputParams) (<-chan *tbPlugin.PrTaxBracketOutputParams, error) {
	out := make(chan *tbPlugin.PrTaxBracketOutputParams)
	go func() {
		defer close(out)
		p.prBracketsMu.Lock()
		defer p.prBracketsMu.Unlock()
		options := &sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true}
		transaction := p.Conn.Begin(options)
		defer transaction.Commit()
		var output *tbPlugin.PrTaxBracketOutputParams
		for _, param := range params {
			cacheKey := prBracketCache{Year: param.Year, Province: param.Province}
			if value, err := p.prBracketCache.Read(cacheKey); err == nil {
				output = value.(*tbPlugin.PrTaxBracketOutputParams)
			} else {
				rows := make([]tables.CanadaTaxBracketsModel, 0)
				transaction.Model(tables.CanadaTaxBracketsModel{}).
					Where("year = ?", param.Year).
					Where("region_code = ?", param.Province).Order("min_salary ASC").Find(&rows)
				entities := make([]*tbPlugin.TaxBracketEntity, 0, len(rows))
				for _, row := range rows {
					entities = append(entities, row.ToTaxBracketEntity())
				}
				output = &tbPlugin.PrTaxBracketOutputParams{
					PrTaxBracketsInputParams: *param,
					Entities:                 entities,
				}
				p.prBracketCache.Add(cacheKey, output)
			}
			out <- output
		}
	}()
	return out, nil
}
