service = sms
package = github.com/liuliqiang/smsToEmail
version = $(shell git describe --long --tags --dirty | awk '{print substr($$1,2)}')
build_dir = _build
go_img = golang:1.20
GOARCH ?= amd64


.PHONY: build docker runforever stop
build:
	go build -o $(name) ./main.go

docker:
	echo "TODO"

release_dir = $(build_dir)/release
release_name = $(service)-$(version).linux-$(GOARCH)
.PHONY: release
release:
	mkdir -p $(release_dir)
	GOOS=linux GOARCH=$(GOARCH) go build -ldflags '-X main.version=$(version)' \
		-gcflags=-trimpath='$(shell pwd)' -o $(release_dir)/$(release_name)/$(service) \
		*.go
	tar -czf $(release_dir)/$(release_name).tar.gz -C $(release_dir) $(release_name)

.PHONY: docker-release
docker-release:
	docker run --privileged --rm -v $(shell pwd):/go/src/$(package) \
		-w /go/src/$(package) -e GOARCH=$(GOARCH) $(go_img) make release

rpm_target = x86_64
rpm_dir = $(build_dir)/rpm
rpm_version = $(shell echo $(version) | sed -nr "s/^([0-9]+(\.[0-9]+)+)(-([-A-Za-z0-9\.]+))?$$/\1/p")
rpm_release = $(subst -,.,$(shell echo $(version) | sed -nr "s/^([0-9]+(\.[0-9]+)+)(-([-A-Za-z0-9\.]+))?$$/\4/p"))
.PHONY: docker-rpm
docker-rpm:
	mkdir -p $(rpm_dir)
	cp build/rpm/* $(rpm_dir)
	cp $(release_dir)/$(release_name).tar.gz $(rpm_dir)
	git log --format="* %cd %aN%n- (%h) %s%d%n" -n 10 --date local \
		| sed -r 's/[0-9]+:[0-9]+:[0-9]+ //' >> $(rpm_dir)/sms.spec
	chmod -R g+w,o+w $(rpm_dir)
	chown -R 1000:1000 $(rpm_dir)
	docker run --privileged --rm -v $(shell pwd)/$(rpm_dir):/home/builder/rpm \
		-w /home/builder/rpm rpmbuild/centos7 \
		rpmbuild --target '$(rpm_target)' --define '_name $(service)' --define '_version $(rpm_version)' \
		--define '_release $(rpm_release)' --define '_source $(release_name)' -bb sms.spec

image:
	docker build -t $(service):$(version) .

clean:
	rm -rf $(build_dir)

runforever: build
	mkdir -p ./logs
	nohup ./$(name) > ./logs/info.log 2>./logs/error.log &

stop:
	killall $(name)
