FROM alpine:3

ADD build/sftp-reporter /sftp-reporter

ENTRYPOINT [ "/sftp-reporter" ]
CMD []