package api

// API describes the set of methods offered over the RPC interface
type API struct {
	NameSpace string      // namespace under which the rpc methods of Service are exposed
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}
