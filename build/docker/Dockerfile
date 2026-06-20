FROM busybox:1.37.0

ARG COMPONENT

WORKDIR /data

COPY bin/chainnet-${COMPONENT} /usr/local/bin/chainnet
COPY --chmod=755 build/docker/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chown 65532:65532 /data

EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
