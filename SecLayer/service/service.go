package service

func Run() (err error) {
	//秒杀请求读取、写入、处理协程
	SecProcessFunc()

	return
}
