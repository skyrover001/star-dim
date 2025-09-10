const axios = require('axios');
const FormData = require('form-data');
const fs = require('fs');
const path = require('path');

// 尝试加载登录配置
function loadConfig() {
    const configPath = path.join(__dirname, 'login_config.js');
    const exampleConfigPath = path.join(__dirname, 'login_config.example.js');
    
    if (fs.existsSync(configPath)) {
        console.log('✅ 使用配置文件: login_config.js');
        return require('./login_config.js');
    } else {
        console.log('⚠️  未找到 login_config.js 文件，使用示例配置');
        console.log('💡 请复制 login_config.example.js 为 login_config.js 并填入真实的登录信息');
        return require('./login_config.example.js');
    }
}

class FileAPITester {
    constructor(baseURL, sessionKey = null) {
        this.baseURL = baseURL;
        this.sessionKey = sessionKey;
        this.headers = {
            'sessionKey': sessionKey,
            'Content-Type': 'application/json'
        };
    }

    // 更新会话密钥
    updateSessionKey(sessionKey) {
        this.sessionKey = sessionKey;
        this.headers['sessionKey'] = sessionKey;
    }

    // 用户登录
    async login(cluster, username, password, host, port = '22') {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/login/`, {
                cluster: cluster,
                username: username,
                password: password,
                host: host,
                port: port
            }, {
                headers: {
                    'Content-Type': 'application/json'
                }
            });
            console.log('登录成功:', response.data);
            
            // 自动更新会话密钥
            if (response.data.session_key) {
                this.updateSessionKey(response.data.session_key);
            }
            
            return response.data;
        } catch (error) {
            console.error('登录失败:', error.response?.data || error.message);
            throw error;
        }
    }

    // 用户登出
    async logout() {
        try {
            if (!this.sessionKey) {
                console.error('没有有效的会话密钥');
                return;
            }
            
            const response = await axios.get(`${this.baseURL}/api/v2/logout/`, {
                headers: { 'session_key': this.sessionKey }
            });
            console.log('登出成功:', response.data);
            
            // 清除会话密钥
            this.updateSessionKey(null);
            
            return response.data;
        } catch (error) {
            console.error('登出失败:', error.response?.data || error.message);
        }
    }

    // 列出目录文件
    async listFiles(cluster, path) {
        try {
            const response = await axios.get(`${this.baseURL}/api/v2/files/`, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey }
            });
            console.log('列出文件成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('列出文件失败:', error.response?.data || error.message);
        }
    }

    // 上传文件
    async uploadFile(cluster, path, filePath, offset = 0, update = false) {
        console.log("this.sessionKey:", this.sessionKey);
        try {
            const formData = new FormData();
            formData.append('cluster', cluster);
            formData.append('path', path);
            formData.append('offset', offset.toString());
            formData.append('update', update.toString());
            formData.append('file', fs.createReadStream(filePath));

            const response = await axios.post(`${this.baseURL}/api/v2/files/transmission/`, formData, {
                headers: {
                    'sessionKey': this.sessionKey,
                    ...formData.getHeaders()
                }
            });
            console.log('上传文件成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('上传文件失败:', error.response?.data || error.message);
        }
    }

    // 下载文件
    async downloadFile(cluster, path, outputPath) {
        try {
            const response = await axios.get(`${this.baseURL}/api/v2/files/download/`, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey },
                responseType: 'stream'
            });

            const writer = fs.createWriteStream(outputPath);
            response.data.pipe(writer);

            return new Promise((resolve, reject) => {
                writer.on('finish', () => {
                    console.log('下载文件成功:', outputPath);
                    resolve();
                });
                writer.on('error', reject);
            });
        } catch (error) {
            console.error('下载文件失败:', error.response?.data || error.message);
        }
    }

    // 获取文件属性
    async getFileAttr(cluster, path) {
        try {
            const response = await axios.get(`${this.baseURL}/api/v2/files/attr/`, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey }
            });
            console.log('获取文件属性成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('获取文件属性失败:', error.response?.data || error.message);
        }
    }

    // 重命名文件
    async renameFile(oldPath, newPath) {
        try {
            const response = await axios.put(`${this.baseURL}/api/v2/files/`, {
                old_path: oldPath,
                new_path: newPath
            }, { headers: this.headers });
            console.log('重命名文件成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('重命名文件失败:', error.response?.data || error.message);
        }
    }

    // 创建文件/目录
    async createFile(path, type) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/`, {
                path: path,
                type: type // 'file' or 'dir'
            }, { headers: this.headers });
            console.log('创建成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('创建失败:', error.response?.data || error.message);
        }
    }

    // 删除文件/目录
    async deleteFile(path, type) {
        try {
            const response = await axios.delete(`${this.baseURL}/api/v2/files/`, {
                data: { path: path, type: type },
                headers: this.headers
            });
            console.log('删除成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('删除失败:', error.response?.data || error.message);
        }
    }

    // 复制文件
    async copyFile(srcPath, dstPath) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/copy/`, {
                src_path: srcPath,
                dst_path: dstPath
            }, { headers: this.headers });
            console.log('复制成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('复制失败:', error.response?.data || error.message);
        }
    }

    // 移动文件
    async moveFile(srcPath, dstPath) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/move/`, {
                src_path: srcPath,
                dst_path: dstPath
            }, { headers: this.headers });
            console.log('移动成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('移动失败:', error.response?.data || error.message);
        }
    }

    // 读取文件内容
    async readFileContent(cluster, path) {
        try {
            const response = await axios.get(`${this.baseURL}/api/v2/files/content/`, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey }
            });
            console.log('读取文件内容成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('读取文件内容失败:', error.response?.data || error.message);
        }
    }

    // 写入文件内容
    async writeFileContent(cluster, path, content) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/content/`, {
                cluster: cluster,
                path: path,
                content: content
            }, { headers: this.headers });
            console.log('写入文件内容成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('写入文件内容失败:', error.response?.data || error.message);
        }
    }

    // 执行脚本文件
    async executeFile(cluster, path) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/execute/`, null, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey }
            });
            console.log('执行文件成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('执行文件失败:', error.response?.data || error.message);
        }
    }

    // 修改文件权限
    async changeMode(path, mode) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/chmod/`, {
                path: path,
                mode: mode
            }, { headers: this.headers });
            console.log('修改权限成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('修改权限失败:', error.response?.data || error.message);
        }
    }

    // 修改文件所有者
    async changeOwner(path, owner, group) {
        try {
            const response = await axios.post(`${this.baseURL}/api/v2/files/chown/`, {
                path: path,
                owner: owner,
                group: group
            }, { headers: this.headers });
            console.log('修改所有者成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('修改所有者失败:', error.response?.data || error.message);
        }
    }

    // 获取配额信息
    async getQuota(cluster, path = '') {
        try {
            const response = await axios.get(`${this.baseURL}/api/v2/files/quota/`, {
                params: { cluster, path },
                headers: { 'sessionKey': this.sessionKey }
            });
            console.log('获取配额信息成功:', response.data);
            return response.data;
        } catch (error) {
            console.error('获取配额信息失败:', error.response?.data || error.message);
        }
    }
}

// 加载配置
const CONFIG = loadConfig();

// 单独的测试函数
async function testLogin(tester) {
    console.log('\n=== 测试登录 ===');
    const loginResult = await tester.login(CONFIG.ssh.cluster, CONFIG.ssh.username, 
                                          CONFIG.ssh.password, CONFIG.ssh.host, CONFIG.ssh.port);
    console.log('登录结果:', loginResult);
    return loginResult;
}

async function testLogout(tester) {
    console.log('\n=== 测试登出 ===');
    await tester.logout();
}

async function testListFiles(tester) {
    console.log('\n=== 测试列出文件 ===');
    await tester.listFiles(CONFIG.ssh.cluster, CONFIG.testPaths?.homeDir || '/home');
}

async function testCreateDir(tester) {
    console.log('\n=== 测试创建目录 ===');
    await tester.createFile(CONFIG.testPaths?.testDir || '/home/test_dir', 'dir');
}

async function testCreateFile(tester) {
    console.log('\n=== 测试创建文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.createFile(`${testDir}/test.txt`, 'file');
}

async function testWriteFileContent(tester) {
    console.log('\n=== 测试写入文件内容 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.writeFileContent(CONFIG.ssh.cluster, `${testDir}/test.txt`, 'Hello World from API Test!');
}

async function testReadFileContent(tester) {
    console.log('\n=== 测试读取文件内容 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.readFileContent(CONFIG.ssh.cluster, `${testDir}/test.txt`);
}

async function testGetFileAttr(tester) {
    console.log('\n=== 测试获取文件属性 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.getFileAttr(CONFIG.ssh.cluster, `${testDir}/test.txt`);
}

async function testChangeMode(tester) {
    console.log('\n=== 测试修改权限 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.changeMode(`${testDir}/test.txt`, '755');
}

async function testCopyFile(tester) {
    console.log('\n=== 测试复制文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.copyFile(`${testDir}/test.txt`, `${testDir}/test_copy.txt`);
}

async function testRenameFile(tester) {
    console.log('\n=== 测试重命名文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.renameFile(`${testDir}/test_copy.txt`, `${testDir}/test_renamed.txt`);
}

async function testMoveFile(tester) {
    console.log('\n=== 测试移动文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.moveFile(`${testDir}/test.txt`, `${testDir}/moved_test.txt`);
}

async function testDeleteFile(tester) {
    console.log('\n=== 测试删除文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.deleteFile(`${testDir}/test_renamed.txt`, 'file');
}

async function testUploadFile(tester) {
    console.log('\n=== 测试上传文件 ===');
    await tester.uploadFile(CONFIG.ssh.cluster, '/pic.png', './pic.png', 0, true);
}

async function testDownloadFile(tester) {
    console.log('\n=== 测试下载文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.downloadFile(CONFIG.ssh.cluster, `${testDir}/test.txt`, './downloaded_test.txt');
}

async function testExecuteFile(tester) {
    console.log('\n=== 测试执行文件 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.executeFile(CONFIG.ssh.cluster, `${testDir}/test.sh`);
}

async function testChangeOwner(tester) {
    console.log('\n=== 测试修改所有者 ===');
    const testDir = CONFIG.testPaths?.testDir || '/home/test_dir';
    await tester.changeOwner(`${testDir}/test.txt`, 'root', 'root');
}

async function testGetQuota(tester) {
    console.log('\n=== 测试获取配额信息 ===');
    await tester.getQuota(CONFIG.ssh.cluster);
}

// 接口映射表
const API_TESTS = {
    'login': testLogin,
    'logout': testLogout,
    'list': testListFiles,
    'listFiles': testListFiles,
    'createDir': testCreateDir,
    'createFile': testCreateFile,
    'writeContent': testWriteFileContent,
    'readContent': testReadFileContent,
    'getAttr': testGetFileAttr,
    'chmod': testChangeMode,
    'copy': testCopyFile,
    'rename': testRenameFile,
    'move': testMoveFile,
    'delete': testDeleteFile,
    'upload': testUploadFile,
    'download': testDownloadFile,
    'execute': testExecuteFile,
    'chown': testChangeOwner,
    'quota': testGetQuota
};

// 显示可用的测试接口
function showAvailableTests() {
    console.log('\n可用的测试接口:');
    Object.keys(API_TESTS).forEach((key, index) => {
        console.log(`  ${index + 1}. ${key}`);
    });
    console.log('\n使用方法:');
    console.log('  node all_test.js --test login                    # 测试单个接口');
    console.log('  node all_test.js --test login,upload,logout     # 测试多个接口');
    console.log('  node all_test.js --login-only                   # 仅登录测试');
    console.log('  node all_test.js                                # 完整测试');
    console.log('  node all_test.js --help                         # 显示帮助');
}

// 运行指定的测试
async function runSpecificTests(testNames) {
    const tester = new FileAPITester(CONFIG.server.baseURL);
    let isLoggedIn = false;
    
    try {
        for (const testName of testNames) {
            const testFunc = API_TESTS[testName];
            if (!testFunc) {
                console.error(`未知的测试接口: ${testName}`);
                continue;
            }
            
            // 如果需要登录且尚未登录
            if (testName !== 'login' && !isLoggedIn) {
                await testLogin(tester);
                isLoggedIn = true;
            }
            
            // 执行测试
            if (testName === 'login') {
                await testFunc(tester);
                isLoggedIn = true;
            } else if (testName === 'logout') {
                await testFunc(tester);
                isLoggedIn = false;
            } else {
                await testFunc(tester);
            }
        }
        
        // 如果已登录但没有执行登出测试，自动登出
        if (isLoggedIn && !testNames.includes('logout')) {
            await testLogout(tester);
        }
        
    } catch (error) {
        console.error('测试过程中出现错误:', error);
    }
}

// 完整测试
async function runTests() {
    const tester = new FileAPITester(CONFIG.server.baseURL);

    try {
        // 登录
        await testLogin(tester);
        
        // 测试上传文件 (保留原有的上传测试)
        await testUploadFile(tester);
        
        // 登出
        await testLogout(tester);

    } catch (error) {
        console.error('测试过程中出现错误:', error);
    }
}

// 单独的登录测试函数
async function testLoginOnly() {
    const tester = new FileAPITester(CONFIG.server.baseURL);
    
    try {
        console.log('开始登录测试...');
        const result = await tester.login('default', CONFIG.ssh.username, 
                                          CONFIG.ssh.password, CONFIG.ssh.host, CONFIG.ssh.port);
        console.log('登录测试成功:', result);
        
        // 测试一个简单的API调用来验证会话
        await tester.listFiles('default', '/');
        
        // 登出
        await tester.logout();
        
    } catch (error) {
        console.error('登录测试失败:', error);
    }
}

// 运行测试
if (require.main === module) {
    const args = process.argv.slice(2);
    
    // 显示帮助信息
    if (args.includes('--help') || args.includes('-h')) {
        showAvailableTests();
        process.exit(0);
    }
    
    // 查找 --test 参数
    const testIndex = args.findIndex(arg => arg === '--test');
    if (testIndex !== -1 && args[testIndex + 1]) {
        // 解析测试接口名称，支持逗号分隔的多个接口
        const testNames = args[testIndex + 1].split(',').map(name => name.trim());
        console.log(`开始运行指定测试: ${testNames.join(', ')}`);
        runSpecificTests(testNames).catch(console.error);
    } else if (args.includes('--login-only')) {
        testLoginOnly().catch(console.error);
    } else if (args.length === 0) {
        console.log('选择测试模式:');
        console.log('1. 完整测试 (包含登录)');
        console.log('2. 仅登录测试');
        console.log('3. 指定测试 (使用 --test 参数)');
        console.log('4. 显示帮助 (使用 --help 参数)');
        runTests().catch(console.error);
    } else {
        console.log('无效的参数，使用 --help 查看帮助信息');
        showAvailableTests();
    }
}

module.exports = { 
    FileAPITester, 
    runTests, 
    testLoginOnly, 
    runSpecificTests,
    API_TESTS,
    showAvailableTests,
    loadConfig,
    CONFIG
};