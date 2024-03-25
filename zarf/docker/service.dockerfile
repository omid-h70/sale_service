# Base go image
# Build in Vendor mod
#
# if you want to use yor args in multiple stages
## first define them and refer to them in stages
##
##
##
ARG APP_PATH=/service
ARG ADMIN_APP_PATH=/service/tooling/admin
ARG BUILD_REF=develop
ARG APP_NAME=service

FROM golang:1.21-alpine as builder

ARG APP_PATH
ARG ADMIN_APP_PATH
ARG BUILD_REF
ARG APP_NAME

# env vars are available on runtimes but args not
ENV APP_NAME=$APP_NAME
ENV ADMIN_APP_NAME=admin


#ARG is only available during the build of a Docker image (RUN etc),
#not after the image is created and containers are started from it (ENTRYPOINT, CMD)

RUN mkdir ${APP_PATH}

COPY . ${APP_PATH}

WORKDIR ${APP_PATH}
## build admin binary
RUN CGO_ENABLED=0 go build -ldflags="-X main.build=${BUILD_REF}" -mod=vendor -o ${APP_NAME}
RUN chmod +x ${APP_NAME}

WORKDIR ${ADMIN_APP_PATH}
## build app binary
RUN CGO_ENABLED=0 go build -ldflags="-X main.build=${BUILD_REF}" -mod=vendor -o ${ADMIN_APP_NAME}
RUN chmod +x ${ADMIN_APP_NAME}

#If you want your CMD to expand variables,
#you need to arrange for a shell. You can do that like below:
#CMD [ "sh", "-c", "./${APP_NAME}" ]


###### Production Image - tiny One !
# It was actually very common to have one Dockerfile to use for development (which contained everything needed to build your application),
# and a slimmed-down one to use for production, which only contained your application and exactly what was needed to run it.
# This has been referred to as the “builder pattern”. Maintaining two Dockerfiles is not ideal.
# then moving the built package to a tiny docker image
#
#
FROM alpine:latest
ARG BUILD_REF
ARG BUILD_DATE
ARG APP_NAME
ARG APP_PATH
#

#COPY --from=builder /service/zarf/keys/ /service/zarf/keys/
#COPY --from=builder /service/app/tooling/admin /service/admin
RUN echo ${APP_PATH}
COPY --from=builder ${APP_PATH} /service/sales-api

WORKDIR /service
#
##RUN CGO_ENABLED=0 go build -mod=vendor -o service
#
##RUN chmod +x $BINARY_PATH/service
#
CMD [ "./service" ]