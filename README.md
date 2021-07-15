Feye

### 前言

初学golang写个爬虫练练手

### Feye说明

用于爬取fofa zoom，用于刷src，手动测试fofa上面每一个链接，并且进行漏洞测试着实麻烦，结合poc快速刷src

，当然也可以封装为package

### 使用

git clone https://github.com/Lmg66/Feye.git

配置config.yaml

![](https://cdn.jsdelivr.net/gh/Lmg66/picture@main/image/1626268360555-9999.png)

 Authorization登录fofa账号，随便查询一个语句，点第二页，burpsuite抓包将 Authorization粘贴到配置文件即可

![](https://cdn.jsdelivr.net/gh/Lmg66/picture@main/image/1626268935029-9999967.png)

### Windows

![](https://cdn.jsdelivr.net/gh/Lmg66/picture@main/image/1626311997219-212542.png)

### Linux

![](https://cdn.jsdelivr.net/gh/Lmg66/picture@main/image/1626312008362-212859.png)

### Mac

无mac无测试

