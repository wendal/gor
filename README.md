## Table of Contents
* [English Introduction](#english-introduction)
    * [Installation](#installation)
    * [Quick Start](#quick-start)
* [中文介绍](#chinese-introduction)
    * [安装](#installation-)
    * [快速入门](#quick-start-)
* [Copyright and License](#copyright-and-license)
* [Sites using gor](#sites-using-gor)

# English Introduction

## gor -- A static websites and blog generator engine written in Go
===

Transform your plain text into static websites and blogs.
gor is a [Ruhoh](http://ruhoh.com/) like websites and blog generator engine written in [Go](http://golang.org/). It's almost compatible to ruhoh 1.x specification. You can treat gor as a replacement of the official implementation what is written in [Ruby](http://www.ruby-lang.org/en/). 

Why reinvent a wheel? gor has following awesome benefits:
1. Speed -- Less than 1 second when compiling all my near 200 blogs on wendal.net
2. Simple -- Only one single executable file generated after compiling, no other dependence

## Installation
====================
To install:

    go get -u github.com/wendal/gor
    go install github.com/wendal/gor/gor


If you use [brew](https://github.com/mxcl/homebrew) on Mac, and you didn't set `$GOROOT` and `$GOPATH` environment variable
Please using this command:
		ln -s /usr/local/Cellar/go/1.0.3/bin/gor /usr/local/bin
	
Or to download a compiled one directly from googecode[gor.zip](https://code.google.com/p/gor/downloads/list)

## Quick Start
======================

Create a new website
-------

	gor new example.com
	cd example.com
    #After execution, a folder named example.com will be generated, including a scaffold & some sample posts.

Create a new post
----------

	gor post "goodday" [dir/to/img/files]
	#generate a new post file: post/goodday.md, open it with your markdown editor to write.

`dir/to/img/files` is optionl. If it's provided, all files in that dir will be copy into blog dir(configurable dir), and insert `<img>` tag into post file. 

Configuration
--------

Open the site.yml file in root folder

1. Input title, author etc.
2. Input email etc.

Open the config.yml file in root folder

1. production_url is your website address, such as http://wendal.net, don't add '/' at last, it will be used to generate rss.xml etc.
2. summary_lines is the length of abstract on homepage, any number as you like.
3. latest is how many posts will be shown on homepage
4. `imgs` parts is auto img config
   * `imgtag`：basic format for <img> tag to be insert. the `%s` part will to replaced by `urlperfix/post_name/img_file_name`
   * `urlperfix`：img file url perfix
   * `localdir`：location inside blog repo for img file storage

Open widgets folder, you can see some widgets here, there is a config.yml file of each widget for configuration.

1. analytics only support google analytics by now, please input tracking_id here
2. comments only support disqus by now, please input your short_name of disqus here
3. google_prettify for code highlighting, normally it's not necessary to change


Compile to generate static web page
--------------

	gor compile
	#Finished instantly. A new folder named compiled will be generated, all website is in it.

Local preview
-------
gor also comes with a built-in development server that will allow you to preview what the generated site will look like in your browser locally.

	gor http
	#Open your favorite web browser and visit: http://127.0.0.1:8080

Deployment
-----

You can deploy it to [GitHub Pages](http://pages.github.com/), or put it to your own VPS, because there are only static files(HTML, CSS, js etc.), no need of php/mysql/java etc. 

# Chinese Introduction

## gor -- Golang编写的静态博客引擎
===

gor 是使用golang实现的类Ruhoh静态博客引擎(Ruhoh like),基本兼容ruhoh 1.x规范.
相当于与ruhoh的官方实现(ruby实现), 有以下优点:

1. 速度完胜 -- 编译wendal.net近200篇博客,仅需要1秒
2. 安装简单 -- 得益于golang的特性,编译后仅一个可运行程序,无依赖

## Installation 安装
====================
To install:

    go get -u github.com/wendal/gor
    go install github.com/wendal/gor/gor

在Mac下使用brew的用户

如果是通过[brew](https://github.com/mxcl/homebrew)来安装`go`，并且没有设置`$GOROOT`跟`$GOPATH`的话，请使用如下命令

    ln -s /usr/local/Cellar/go/1.0.3/bin/gor /usr/local/bin
	
或者你可以从googecode直接下载[编译好的gor](http://gor.googlecode.com)

## Quick Start 快速入门
======================

新建站点
-------

	gor new example.com
    #执行完毕后, 会生成example.com文件夹,包含基本素材及演示文章

新建单篇博客
----------

	cd example.com
	gor post "goodday" [dir/to/img/files]
	#即可生成 post/goodday.md文件, 打开你的markdown编辑器即可编写

如果输入可选参数 `dir/to/img/files`，gor 会从该目录拷贝图片文件到配置的目录，同时在 `goodday.md` 中自动插入对应的 `<img>` 标签。

基本配置
--------

打开站点根目录下的site.yml文件

1. 填入title, 作者等信息
2. 填入邮箱等信息

打开站点根目录下的config.yml文件

1. 设置production_url为你的网站地址, 例如 http://wendal.net 最后面不需要加入/ 生成rss.xml等文件时会用到
2. summary_lines 首页的文章摘要的长度,按你喜欢的呗
3. latest 首页显示多少文章
4. imgs 部分为自动插入 `<img>` 相关的配置
   * `imgtag`：要插入的 <img> 标签的基本格式，`%s` 部分会被自动替换为 `urlperfix/post_name/img_file_name` 的格式
   * `urlperfix`： 图片地址前缀
   * `localdir`：图片文件在博客内的本地存放目录

打开widgets目录, 可以看到基本的挂件,里面有config.yml配置文件

1. analytics 暂时只支持google analytics, 填入tracking_id
2. comments 暂时只支持disqus, 请填入short_name
3. google_prettify 代码高亮,一般不修改


编译生成静态网页
--------------

	gor compile
	#瞬间完成,生成compiled文件夹,包含站点所有资源

本地预览
-------

	gor http
	#打开你的浏览器,访问 http://127.0.0.1:8080

部署
-----

你可以使用github pages等服务,或者放到你的自己的vps下, 因为是纯静态文件,不需要php/mysql/java等环境的支持


Copyright and License
----------------------

This project is licensed under the BSD license.

	
Copyright (C) 2013, by WendalChen wendal1985@gmail.com.

Sites using gor
-----------------------
正在使用Gor的博客
-----------------------

[When Go meets Raspberry Pi](http://hugozhu.myalert.info/)

[RaymondChou's Blog](http://ledbk.com/)

[visualfc's blog](http://visualfc.github.com/)

[wendal随笔](http://wendal.net)

[一个热爱木工的软件工程师](http://iwood.co.nr)

If you are also using gor, please don't hesitate to tell me by email or open an issue.
如果也在使用,欢迎email或者开个issue告诉我们哦
