FROM iron/base

COPY ./drogo ./drogo

ENV ES_HOST 139.59.85.55
ENV ES_PORT 9200
ENV ES_INDEX events
ENV ES_INDEX_TYPE event

ENV ARANGO_HOST=http://139.59.85.55:8529
ENV ARANGO_DB=eventackle
ENV ARANGO_USERNAME=root
ENV ARANGO_PASSWORD=qF3mKQcu7zyzBYly

EXPOSE 80

CMD ["./drogo", "roar"]