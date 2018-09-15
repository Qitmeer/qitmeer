# nox


##  Prerequisites

- Update Go to latest version 1.11

```bash
~ go version
go version go1.11 darwin/amd64
```
## How to build

- 添加以下字段到你的环境变量  
（$HOME/.zshrc $HOME/.bashrc $HOME/.bash_profile /etc/profile等选择合适的）

```bash
export GO111MODULE=auto
```

```bash
~ mkdir -p /tmp/work
~ cd /tmp/work
~ git clone https://github.com/noxproject/nox 
~ git checkout cleanup 
~ go build
```

***注意：一旦启用go module，就尽可能不要把项目clone到GOPATH之下，会造成依赖的问题***

### Go Mod

#### 提示
> download : download modules to local cache (下载依赖的module到本地cache))  
> edit : edit go.mod from tools or scripts (编辑go.mod文件)  
> graph : print module requirement graph (打印模块依赖图))  
> init : initialize new module in current directory (再当前文件夹下初始化一个新的module, 创建go.mod文件))  
> tidy : add missing and remove unused modules (增加丢失的module，去掉未用的module)  
> vendor : make vendored copy of dependencies (将依赖复制到vendor下)  
> verify : verify dependencies have expected content (校验依赖)  
> why : explain why packages or modules are needed (解释为什么需要依赖)  

```bash
~ go mod tidy
~ go mod verify
~ go build
```

```bash
~ go build
~ ./nox --version
```
---

**happy hacking!**

