package middleware

import "net/http"

// RouterGroup 给一组路由统一挂载拦截器
type RouterGroup struct {
	Mux          *http.ServeMux
	Interceptors []Interceptor // 分组拦截器
}

// NewRouterGroup 创建 RouterGroup
func NewRouterGroup(mux *http.ServeMux, interceptors ...Interceptor) *RouterGroup {
	return &RouterGroup{
		Mux:          mux,
		Interceptors: interceptors,
	}
}

// Handle 注册路由，同时应用分组拦截器 + 单条路由拦截器
func (g *RouterGroup) Handle(path string, handler http.HandlerFunc, interceptors ...Interceptor) {
	all := append(g.Interceptors, interceptors...)
	g.Mux.HandleFunc(path, Chain(handler, all...))
}

// RouterManager 管理全局 + 分组
type RouterManager struct {
	Mux                *http.ServeMux
	GlobalInterceptors []Interceptor
	Groups             []*RouterGroup
}

// NewRouterManager 创建 RouterManager
func NewRouterManager(mux *http.ServeMux, globals ...Interceptor) *RouterManager {
	return &RouterManager{
		Mux:                mux,
		GlobalInterceptors: globals,
		Groups:             []*RouterGroup{},
	}
}

// NewGroup 创建分组，自动应用全局拦截器
func (rm *RouterManager) NewGroup(interceptors ...Interceptor) *RouterGroup {
	group := &RouterGroup{
		Mux:          rm.Mux,
		Interceptors: append(rm.GlobalInterceptors, interceptors...),
	}
	rm.Groups = append(rm.Groups, group)
	return group
}

// Handle 注册单条路由，应用全局拦截器 + 单条拦截器
func (rm *RouterManager) Handle(path string, handler http.HandlerFunc, interceptors ...Interceptor) {
	all := append(rm.GlobalInterceptors, interceptors...)
	rm.Mux.HandleFunc(path, Chain(handler, all...))
}
