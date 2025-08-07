const { FileAPITester } = require('./all_test.js');
const fs = require('fs');
const path = require('path');

// 尝试加载登录配置
function loadConfig() {
    const configPath = path.join(__dirname, 'login_config.js');
    const exampleConfigPath = path.join(__dirname, 'login_config.example.js');
    
    if (fs.existsSync(configPath)) {
        return require('./login_config.js');
    } else {
        console.log('未找到 login_config.js 文件');
        console.log('请复制 login_config.example.js 为 login_config.js 并填入真实的登录信息');
        return require('./login_config.example.js');
    }
}

// 带配置的登录测试
async function testWithConfig() {
    const config = loadConfig();
    const tester = new FileAPITester(config.server.baseURL);
    
    try {
        console.log('=== 开始登录测试 ===');
        console.log(`连接服务器: ${config.server.baseURL}`);
        console.log(`SSH主机: ${config.ssh.host}:${config.ssh.port}`);
        console.log(`用户: ${config.ssh.username}`);
        
        // 登录
        const loginResult = await tester.login(
            config.ssh.cluster,
            config.ssh.username,
            config.ssh.password,
            config.ssh.host,
            config.ssh.port
        );
        
        console.log('✅ 登录成功!');
        console.log('会话密钥:', loginResult.session_key);
        console.log('主目录:', loginResult.home_path);
        
        // 测试基本文件操作
        console.log('\n=== 测试基本文件操作 ===');
        
        // 列出主目录文件
        console.log('📁 列出主目录文件...');
        await tester.listFiles(config.ssh.cluster, loginResult.home_path || config.testPaths.homeDir);
        
        // 创建测试目录
        console.log('📁 创建测试目录...');
        const testDir = config.testPaths.testDir;
        await tester.createFile(testDir, 'dir');
        
        // 创建测试文件
        console.log('📄 创建测试文件...');
        const testFile = `${testDir}/login_test.txt`;
        await tester.createFile(testFile, 'file');
        
        // 写入内容
        console.log('✏️  写入文件内容...');
        const content = `登录测试成功!\n时间: ${new Date().toISOString()}\n会话密钥: ${loginResult.session_key}`;
        await tester.writeFileContent(config.ssh.cluster, testFile, content);
        
        // 读取内容验证
        console.log('📖 读取文件内容验证...');
        await tester.readFileContent(config.ssh.cluster, testFile);
        
        // 清理测试文件
        console.log('🧹 清理测试文件...');
        await tester.deleteFile(testFile, 'file');
        await tester.deleteFile(testDir, 'dir');
        
        // 登出
        console.log('\n=== 登出 ===');
        await tester.logout();
        console.log('✅ 登出成功!');
        
        console.log('\n🎉 所有测试完成!');
        
    } catch (error) {
        console.error('❌ 测试失败:', error.message);
        if (error.response?.data) {
            console.error('服务器响应:', error.response.data);
        }
    }
}

// 简单的连接测试
async function testConnection() {
    const config = loadConfig();
    const tester = new FileAPITester(config.server.baseURL);
    
    try {
        console.log(`🔗 测试服务器连接: ${config.server.baseURL}`);
        
        // 尝试登录
        const loginResult = await tester.login(
            config.ssh.cluster,
            config.ssh.username,
            config.ssh.password,
            config.ssh.host,
            config.ssh.port
        );
        
        console.log('✅ 连接成功!');
        console.log('会话密钥:', loginResult.session_key);
        
        // 立即登出
        await tester.logout();
        console.log('✅ 登出成功!');
        
    } catch (error) {
        console.error('❌ 连接失败:', error.message);
        if (error.response?.data) {
            console.error('错误详情:', error.response.data);
        }
    }
}

// 命令行参数处理
if (require.main === module) {
    const args = process.argv.slice(2);
    
    if (args.includes('--connection-only')) {
        console.log('执行连接测试...\n');
        testConnection().catch(console.error);
    } else {
        console.log('执行完整测试...\n');
        testWithConfig().catch(console.error);
    }
}

module.exports = { testWithConfig, testConnection, loadConfig };
