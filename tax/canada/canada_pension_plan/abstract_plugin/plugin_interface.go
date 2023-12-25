package abstract_plugin

type CppService interface {
	GetCppEntities(params []*CppPluginInputParameters) (<-chan *CppEntity, error)
}
