if [ $1 = 1 ]; then
    /sbin/chkconfig --add prometheus-aggregate-exporter
    chown prometheus:prometheus /etc/init.d/prometheus-aggregate-exporter
    chmod 755 /etc/init.d/prometheus-aggregate-exporter
    chown -R prometheus:prometheus /etc/sysconfig/prometheus
    chown -R prometheus:prometheus /etc/sysconfig/logrotate.d
    chown prometheus:prometheus /var/log
    chown prometheus:prometheus /var/run
    chmod 755 /etc/init.d/prometheus-aggregate-exporter
    chmod -R 755 /etc/sysconfig/prometheus
    chmod -R 755 /etc/sysconfig/logrotate.d
fi
