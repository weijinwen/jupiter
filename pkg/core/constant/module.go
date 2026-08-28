package constant

type Module int

const (
	ModuleInvalid = Module(iota)

	ModuleClientGrpc
	ModuleClientResty
	ModuleClientRedis
	ModuleClientRocketMQ
	ModuleClientEtcd
	ModuleClientConsul
	ModuleClientKafka

	ModuleRegistryEtcd
	ModuleRegistryConsul

	ModuleStoreMongoDB
	ModuleStoreGorm
	ModuleStoreTableStore
	ModuleClusterRedis
	ModuleClientSearch
)
