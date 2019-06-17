### Boot's application
* apiVersion: app.logancloud.com/v1
* kind: 
    - JavaBoot: For java application
    - PhpBoot: For php application
    - PythonBoot: For python application
    - NodeJSBoot: For nodejs(runs with nodejs) application
    - WebBoot: For web(runs with nginx) application
    
### Boot's spec properties
Currently, only Image and Version is required, other properties could use the global default.

- Image：Image。**require**
- Version：Image version。**require**
- Replicas：application replicas
- Env：application's environment
- Port：application's listen port
- SubDomain：for reservation, not use now
- Resources：application's resource
- Health：application's health check url
- NodeSelector：application's nodeSelector 
- Command: the command for application's container, override the image.
    
### Middleware(TODO)
* apiVersion: middleware.logancloud.com/v1
* kind: 
    - kafka
    - memcached    