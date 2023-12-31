package goi

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
)

type AsView struct {
	GET     HandlerFunc
	HEAD    HandlerFunc
	POST    HandlerFunc
	PUT     HandlerFunc
	PATCH   HandlerFunc
	DELETE  HandlerFunc
	CONNECT HandlerFunc
	OPTIONS HandlerFunc
	TRACE   HandlerFunc
}

type routerParam struct {
	paramName string
	paramType string
}

// 路由表
type metaRouter struct {
	includeRouter map[string]*metaRouter // 子路由
	viewSet       AsView                 // 视图方法
	staticPattern string                 // 是否是静态路由
}

// 创建路由
func newRouter() *metaRouter {
	return &metaRouter{
		includeRouter: make(map[string]*metaRouter),
		viewSet:       AsView{},
		staticPattern: "",
	}
}

// 添加父路由
func (router *metaRouter) Include(UrlPath string) *metaRouter {
	if _, ok := router.includeRouter[UrlPath]; ok {
		panic(fmt.Sprintf("路由已存在: %s\n", UrlPath))
	}
	var re *regexp.Regexp
	for includePatternUri, Irouter := range router.includeRouter {
		if len(Irouter.includeRouter) == 0 && Irouter.staticPattern == "" {
			re = regexp.MustCompile(includePatternUri + "%")
		} else {
			re = regexp.MustCompile(includePatternUri + "/") // 拥有子路由,或静态路径
		}
		if len(re.FindStringSubmatch(UrlPath)) != 0 {
			panic(fmt.Sprintf("%s 中包含的子路由已被注册: %s\n", UrlPath, includePatternUri))
		}
	}
	router.includeRouter[UrlPath] = &metaRouter{
		includeRouter: make(map[string]*metaRouter),
		viewSet:       AsView{},
		staticPattern: "",
	}
	return router.includeRouter[UrlPath]
}

// 添加路由
func (router *metaRouter) UrlPatterns(UrlPath string, view AsView) {
	if _, ok := router.includeRouter[UrlPath]; ok {
		panic(fmt.Sprintf("路由已存在: %s\n", UrlPath))
	}
	var re *regexp.Regexp
	for includePatternUri, Irouter := range router.includeRouter {
		if len(Irouter.includeRouter) == 0 && Irouter.staticPattern == "" {
			re = regexp.MustCompile(includePatternUri + "%")
		} else {
			re = regexp.MustCompile(includePatternUri + "/") // 拥有子路由,或静态路径
		}
		if len(re.FindStringSubmatch(UrlPath)) != 0 {
			panic(fmt.Sprintf("%s 中的父路由已被注册: %s\n", UrlPath, includePatternUri))
		}
	}

	router.includeRouter[UrlPath] = &metaRouter{
		includeRouter: make(map[string]*metaRouter),
		viewSet:       view,
		staticPattern: "",
	}
}

// 添加静态路由
func (router *metaRouter) StaticUrlPatterns(UrlPath string, StaticUrlPath string) {
	if _, ok := router.includeRouter[UrlPath]; ok {
		panic(fmt.Sprintf("静态映射路由已存在: %s\n", UrlPath))
	}
	var re *regexp.Regexp
	for includePatternUri, Irouter := range router.includeRouter {
		if len(Irouter.includeRouter) == 0 && Irouter.staticPattern == "" {
			re = regexp.MustCompile(includePatternUri + "%")
		} else {
			re = regexp.MustCompile(includePatternUri + "/") // 拥有子路由,或静态路径
		}
		if len(re.FindStringSubmatch(UrlPath)) != 0 {
			panic(fmt.Sprintf("%s 中的父路由已被注册: %s\n", UrlPath, includePatternUri))
		}
	}
	dir, _ := os.Getwd()
	absolutePath := path.Join(dir, StaticUrlPath)
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("静态映射路径不存在: %s\n", StaticUrlPath))
	}

	router.includeRouter[UrlPath] = &metaRouter{
		includeRouter: make(map[string]*metaRouter),
		viewSet:       AsView{GET: metaStaticHandler},
		staticPattern: absolutePath,
	}
}

// 路由器处理程序
func (router *metaRouter) routerHandlers(request *Request) (handlerFunc HandlerFunc, err string) {

	view_methods, isPattern := routeResolution(request.Object.URL.Path, router.includeRouter, request.PathParams)
	if isPattern == false {
		err = fmt.Sprintf("URL NOT FOUND: %s", request.Object.URL.Path)
		return nil, err
	}
	switch request.Object.Method {
	case "GET":
		handlerFunc = view_methods.GET
	case "HEAD":
		handlerFunc = view_methods.HEAD
	case "POST":
		handlerFunc = view_methods.POST
	case "PUT":
		handlerFunc = view_methods.PUT
	case "PATCH":
		handlerFunc = view_methods.PATCH
	case "DELETE":
		handlerFunc = view_methods.DELETE
	case "CONNECT":
		handlerFunc = view_methods.CONNECT
	case "OPTIONS":
		handlerFunc = view_methods.OPTIONS
	case "TRACE":
		handlerFunc = view_methods.TRACE
	default:
		err = fmt.Sprintf("Method NOT FOUND: %s", request.Object.Method)
		return nil, err
	}

	if handlerFunc == nil {
		err = fmt.Sprintf("Method NOT FOUND: %s", request.Object.Method)
		return nil, err
	}
	return handlerFunc, ""
}

// 路由解析
func routeResolution(requestPattern string, includeRouter map[string]*metaRouter, PathParams metaValues) (AsView, bool) {
	var re *regexp.Regexp
	for includePatternUri, router := range includeRouter {
		params, converterPattern := routerParse(includePatternUri)
		if len(params) == 0 { // 无参数直接匹配
			if len(router.includeRouter) == 0 && router.staticPattern == "" {
				re = regexp.MustCompile(includePatternUri + "$")
			} else {
				re = regexp.MustCompile(includePatternUri + "/")
			}
		} else {
			if len(router.includeRouter) == 0 && router.staticPattern == "" {
				re = regexp.MustCompile(converterPattern + "$")
			} else {
				re = regexp.MustCompile(converterPattern + "/")
			}
		}

		paramsSlice := re.FindStringSubmatch(requestPattern)
		if len(paramsSlice)-1 != len(params) || len(paramsSlice) == 0 {
			continue
		}
		paramsSlice = paramsSlice[1:]
		for i := 0; i < len(params); i++ {
			param := params[i]
			value := parseValue(param.paramType, paramsSlice[i])
			PathParams[param.paramName] = append(PathParams[param.paramName], value) // 添加参数
		}
		if router.staticPattern != "" { // 静态路由映射
			PathParams["staticPattern"] = append(PathParams["staticPattern"], router.staticPattern)
			PathParams["staticPath"] = append(PathParams["static"], requestPattern[len(includePatternUri):])
			return router.viewSet, true
		} else if len(router.includeRouter) == 0 { // API
			return router.viewSet, true
		} else { // 子路由
			return routeResolution(requestPattern[len(includePatternUri):], router.includeRouter, PathParams)
		}
	}
	return AsView{}, false
}

// 路由参数解析
func routerParse(UrlPath string) (params []routerParam, converterPattern string) {
	regexpStr := `<([^/<>]+):([^/<>]+)>`
	re := regexp.MustCompile(regexpStr)
	result := re.FindAllStringSubmatch(UrlPath, -1)
	for _, paramsSlice := range result {
		if len(paramsSlice) == 3 {
			converter, ok := metaConverter[paramsSlice[1]]
			if ok == false {
				continue
			}
			re = regexp.MustCompile(paramsSlice[0])
			UrlPath = re.ReplaceAllString(UrlPath, converter)
			converterPattern = UrlPath

			param := routerParam{
				paramName: paramsSlice[2],
				paramType: paramsSlice[1],
			}
			params = append(params, param)
		}
	}
	return
}

// 参数解析
func parseValue(paramType string, paramValue string) any {
	switch paramType {
	case "int":
		value, _ := strconv.Atoi(paramValue)
		return value
	case "uuid":
		value, _ := uuid.Parse(paramValue)
		return value
	default:
		return paramValue
	}
}

func createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	return func(request *Request) any {
		file := request.PathParams.Get("static").(string)
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			return Response{Status: http.StatusNotFound, Data: ""}
		}
		return fileServer.ServeHTTP
	}
}
