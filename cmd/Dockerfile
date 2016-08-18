FROM alpine

RUN apk add --no-cache ca-certificates

COPY pproxies-alpine /bin/pproxies

ENV PORT 9527
ENV CHUNK_SIZE 6
ENV TEST_URL http://httpbin.org/ip
ENV TIMEOUT 5
ENV MAX_ERROR 10
ENV PROXY_NUM 5

EXPOSE $PORT

CMD /bin/pproxies -p $PORT -s $CHUNK_SIZE -u $TEST_URL -t $TIMEOUT -e $MAX_ERROR -n $PROXY_NUM  $DEFAULT --api $API
