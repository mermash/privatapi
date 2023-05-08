COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

include .env
export

.PHONY: mem
mem:
	@echo "after underpress"
	@echo "-- map profile commit=${COMMIT} build_time=${BUILD_TIME}"
	curl "$(HEAP_URL)" -o mem_out 

.PHONY: cpu
cpu:
	@echo "after underpress"
	@echo "-- cpu profile commit=${COMMIT} build_time=${BUILD_TIME}"
	curl "$(CPU_URL)" -o cpu_out

.PHONY: goroutine
goroutine:
	@echo "after underpress"
	@echo "-- cpu profile commit=${COMMIT} build_time=${BUILD_TIME}"
	curl "$(GOROUTINE_URL)" -o goroutine_out

.PHONY: underpress
underpress:
	@echo "-- underpress app $(APP_URL) commit=${COMMIT} build_time=${BUILD_TIME}"
	ab -t 50 -n 10 -c 5 "$(APP_URL)"

.PHONY: build
build:
	@echo "-- build app commit=${COMMIT} build_time=${BUILD_TIME}"
	go build -o privatapi .

.PHONY: show_mem
show_mem:
	@echo "-- show_mem app commit=${COMMIT} build_time=${BUILD_TIME}"
	open -a Safari ./data/mem_ao.svg

.PHONY: show_cpu
show_cpu:
	@echo "-- show_cpu app commit=${COMMIT} build_time=${BUILD_TIME}"
	open -a Safari ./data/cpu.svg

.PHONY: create_svg
create_svg:
	@echo "-- create_svg app commit=${COMMIT} build_time=${BUILD_TIME}"
	docker compose up

.PHONY: drop_image
drop_image:
	@echo "-- drop_image app commit=${COMMIT} build_time=${BUILD_TIME}"
	docker compose down
	docker rmi mermash/go-privatapi-pprof

.PHONY: push
push:
	@echo "-- push to github ${MESSAGE}"
	git add .
	git status
	git commit -m "${MESSAGE}"
	git push
