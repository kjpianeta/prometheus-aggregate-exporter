FROM centos:6
ARG RPM_FILE_NAME
COPY $RPM_FILE_NAME /
RUN set -ex \
    && yum -y update \
    && yum -y install initscripts
RUN set -ex \
    && yum -y install /$RPM_FILE_NAME