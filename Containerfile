FROM scratch

LABEL org.opencontainers.image.source https://github.com/Marco98/mailcow-birthday-daemon
ENTRYPOINT ["/mailcow-birthday-daemon"]

ENV STATEFILE=/data/state.json
VOLUME [ "/data" ]

COPY mailcow-birthday-daemon /mailcow-birthday-daemon
