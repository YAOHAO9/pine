import Pine from 'pine-client'
import { Middleware } from 'pine-client/lib/common'

(async () => {

    const pine = Pine.init()
    await pine.connect(`ws://127.0.0.1:3114?id=${Math.random()}&token=ksYNdrAo`)


    pine.on('connector.onMsg', (data) => {
        console.warn('connector.onMsg', data)
    })

    Pine.ReqTimeOut = 10 * 1000

    const requestDataJSON = { Name: 'JSON request', hahahahah: 18 }


    await pine.fetchCompressMetadata('connector') // 获取与connector服务通信中需要的消息压缩元数据，在访问connector的任何handler前必需先执行此函数。否则客户端可能因无法解析proto消息

    const result1 = await pine.request('connector.handler', requestDataJSON)

    console.log(result1)
    // 中间件1
    const middleware1: Middleware = (data) => {
        if (data.Code === 200) {
            console.warn(data.Message)
            return true // true 继续交由下个中间键处理，如果没有下一个中间键则执行promise resolve函数
        }
        return false // 消息停止传递
    }

    // 中间件2
    const middleware2: Middleware = (data) => {
        if (data.Code.toString().startsWith('4')) {
            throw new Error(JSON.stringify(data))
        }
        return true
    }

    // 加入中间件并发送请求
    const result2 = await pine.request('connector.handler', requestDataJSON, middleware1, middleware2)
    console.log(result2)

})()