FROM busybox
ADD babl_linux_amd64 /bin/babl
RUN chmod +x /bin/babl
ADD babl-server_linux_amd64 /bin/babl-server
RUN chmod +x /bin/babl-server
ADD oom_linux_amd64 /bin/app
RUN chmod +x /bin/app
CMD ["babl-server"]