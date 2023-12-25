package abstract_plugin

type FedEiService interface {
	GetEIEntities([]*FedEiPluginInputParameters) (<-chan *FedEiEntity, error)
}
