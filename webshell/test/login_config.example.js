// 登录配置示例文件
// 请复制此文件为 login_config.js 并填入真实的登录信息

const loginConfig = {
    // 服务器配置
    server: {
        baseURL: 'http://localhost:8080',  // WebShell 服务器地址
        // 其他可选配置
        // baseURL: 'http://192.168.1.100:9090',  // 自定义IP和端口
        // baseURL: 'http://127.0.0.1:8888',      // 本地测试
    },
    
    // SSH连接配置
    ssh: {
        cluster: 'default',    // 集群名称
        username: 'root',      // 用户名
        password: 'your_password_here',  // 密码
        host: '127.0.0.1',     // SSH服务器地址
        port: '22'             // SSH端口
    },
    
    // 测试路径配置
    testPaths: {
        homeDir: '/home',           // 主目录路径
        testDir: '/home/test_api'   // 测试目录路径
    },
    
    // 测试配置
    test: {
        timeout: 5000,              // 请求超时时间 (毫秒)
        retries: 3,                 // 失败重试次数
        cleanupOnExit: true         // 测试结束后是否清理测试文件
    }
};

module.exports = loginConfig;
