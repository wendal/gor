## gor -- Golang编写的静态博客引擎
===

gor 是使用golang实现的类Ruhoh静态博客引擎(Ruhoh like),基本兼容ruhoh 1.x规范.
相当于与ruhoh的官方实现(ruby实现), 有以下优点:

1. 速度完胜 -- 编译wendal.net近200篇博客,仅需要1秒
2. 安装简单 -- 得益于golang的特性,编译后仅一个可运行程序,无依赖

## Installation 安装
====================
To install:

    go get github.com/wendal/gor
    cd $GOPATH/src/github.com/wendal/gor/
    go build gor.go

## Quick Start 快速入门
======================

新建站点
-------

	gor new kkkk.com
    #执行完毕后, 会生成kkk.com文件夹,包含基本素材及演示文章

新建单篇博客
----------

	cd kkkk.com
	gor post "goodday"
	#即可生成 post/goodday.md文件, 打开你的markdown编辑器即可编写

基本配置
--------

打开站点根目录下的site.yml文件

1. 填入title, 作者等信息
2. 填入邮箱等信息

打开站点根目录下的config.yml文件

1. 设置production_url为你的网站地址, 例如 http://wendal.net 最后面不需要加入/ 生成rss.xml等文件时会用到
2. summary_lines 首页的文章摘要的长度,按你喜欢的呗
3. latest 首页显示多少文章

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