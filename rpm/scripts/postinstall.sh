if [ $1 = 1 ]; then
    /sbin/chkconfig --add %{name}
fi
