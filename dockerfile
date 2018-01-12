FROM amd64/alpine
MAINTAINER Daniel Randall <danny_randall@byu.edu>

ARG NAME
ENV name=${NAME}

RUN ["/sbin/apk", "update"]
RUN ["/sbin/apk", "add", "ca-certificates"]

COPY ${name}-bin ${name}-bin 
COPY version.txt version.txt

# add any required files/folders here

ENTRYPOINT ./${name}-bin
EXPOSE 8000
