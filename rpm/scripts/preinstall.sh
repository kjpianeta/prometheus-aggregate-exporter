# If app user does not exist, create
id prometheus >/dev/null 2>&1
if [ $? != 0 ]; then
    /usr/sbin/groupadd -r prometheus >/dev/null 2>&1
    /usr/sbin/useradd -d /var/run/prometheus -r -g prometheus prometheus >/dev/null 2>&1
fi
