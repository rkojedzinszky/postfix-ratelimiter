FROM scratch

LABEL org.opencontainers.image.authors "Richard Kojedzinszky <richard@kojedz.in>"
LABEL org.opencontainers.image.source https://github.com/rkojedzinszky/postfix-ratelimiter

COPY postfix-ratelimiter /

USER 20563

CMD ["/postfix-ratelimiter"]
