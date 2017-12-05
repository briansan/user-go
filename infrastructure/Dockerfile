FROM golang:1.8

# Add deploy key to build container
RUN mkdir /root/.ssh
ADD infrastructure/manageme_deploy_key /root/.ssh/id_rsa

# Pull the code
ENV GO_PACKAGE_NAME github.com/briansan/ManageMeServer
ENV PROJECT_ROOT $GOPATH/src/$GO_PACKAGE_NAME
ENV MANAGEME_DIR $PROJECT_ROOT

RUN mkdir -p $MANAGEME_DIR
ADD vendor $MANAGEME_DIR/vendor
ADD main.go $MANAGEME_DIR
ADD errors $MANAGEME_DIR/errors
ADD model $MANAGEME_DIR/model
ADD api $MANAGEME_DIR/api
ADD www $MANAGEME_DIR/www

# RUN echo "Host git.toptal.com\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config
# RUN git clone git@git.toptal.com:widovic/project-brian-kim.git $PROJECT_ROOT

# Build it
RUN go build -v $GO_PACKAGE_NAME
RUN mv ManageMeServer /bin

# Move the webapp files
RUN mv $MANAGEME_DIR/www /etc

ENTRYPOINT ["ManageMeServer"]
