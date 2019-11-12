FROM debian:stretch-slim

ADD build/sftp-reporter /sftp-reporter

ENTRYPOINT [ "/sftp-reporter" ]
CMD []