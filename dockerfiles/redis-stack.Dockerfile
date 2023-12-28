FROM ubuntu:jammy

RUN apt-get update -qqy --fix-missing
RUN apt-get upgrade -qqy
RUN apt-get install -y dumb-init gdebi-core
ADD ./redis-stack /var/cache/apt/redis-stack/
RUN mkdir -p /data/redis /data/redisinsight
RUN touch /.dockerenv

RUN gdebi -n /var/cache/apt/redis-stack/redis-stack-server*.deb
RUN apt remove -y gdebi
RUN apt autoremove -y
RUN rm -rf /var/cache/apt

COPY ./etc/scripts/entrypoint.sh /entrypoint.sh
RUN chmod a+x /entrypoint.sh

EXPOSE 6379

ENV REDISBLOOM_ARGS ""
ENV REDISEARCH_ARGS ""
ENV REDISJSON_ARGS ""
ENV REDISTIMESERIES_ARGS ""
ENV REDISGRAPH_ARGS ""
ENV REDIS_ARGS ""

CMD ["/entrypoint.sh"]