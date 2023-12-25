package database

import (
	"github.com/gerdooshell/financial/lib/database/connection_config"
	"github.com/gerdooshell/financial/lib/database/database_engine"
	"github.com/gerdooshell/financial/lib/database/database_hosttag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ConnectionPool struct {
	database_hosttag.HostTag
	database_engine.Engine
	Conn *gorm.DB
}

var connections = make(map[struct {
	Engine  database_engine.Engine
	HostTag database_hosttag.HostTag
}]*ConnectionPool)

func NewConnectionPool(config connection_config.ConnectionConfig) (*ConnectionPool, error) {
	if connPool, ok := connections[config.GetSignature()]; ok {
		return connPool, nil
	}
	conn, err := gorm.Open(postgres.Open(config.GetConnectionString()))
	if err != nil {
		return nil, err
	}
	connPool := &ConnectionPool{
		HostTag: config.GetHostTag(),
		Engine:  config.GetEngine(),
		Conn:    conn,
	}
	connections[config.GetSignature()] = connPool
	return connPool, nil
}
