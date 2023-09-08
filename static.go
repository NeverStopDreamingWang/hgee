package hgee

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

// 静态文件处理函数
func metaStaticHandler(request *Request) any {
	staticPattern := request.PathParams.Get("staticPattern").(string)
	staticPath := request.PathParams.Get("staticPath").(string)

	absolutePath := path.Join(staticPattern, staticPath)
	if _, err := os.Stat(absolutePath); os.IsNotExist(err) {
		return Response{
			Status: http.StatusNotFound,
			Data:   nil,
		}
	}
	file, err := os.Open(absolutePath)
	if err != nil {
		file.Close()
		return Response{
			Status: http.StatusInternalServerError,
			Data:   "读取文件失败！",
		}
	}
	return file
}

// 返回响应文件内容
func metaResponseStatic(file *os.File, request *Request, response http.ResponseWriter) (int64, []byte) {
	defer file.Close()

	fileInfo, err := file.Stat()
	if os.IsNotExist(err) {
		response.WriteHeader(http.StatusNotFound)
		return 0, []byte("文件不存在")
	} else if err != nil {
		response.WriteHeader(http.StatusNotFound)
		return 0, []byte(fmt.Sprintf("获取文件状态失败：%v", err))
	}
	if fileInfo.IsDir() {
		response.WriteHeader(http.StatusNotFound)
		return 0, []byte("不是一个文件")
	}
	http.ServeContent(response, request.Object, fileInfo.Name(), time.Now(), file)
	return fileInfo.Size(), nil
}
