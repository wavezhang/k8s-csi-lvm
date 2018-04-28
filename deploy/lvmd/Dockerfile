FROM 172.16.59.153/base/centos:7.2.1511
RUN yum install -y lvm2  device-mapper-persistent-data
COPY ./lvmd /lvmd
CMD ["/lvmd", "-v", "5", "-listen", "0.0.0.0:1736", "-logtostderr"]

