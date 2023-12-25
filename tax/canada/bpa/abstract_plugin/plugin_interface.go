package abstract_plugin

type BPAService interface {
	GetBPAEntities([]*BpaPluginInputParameters) (<-chan *BpaEntity, error)
}
