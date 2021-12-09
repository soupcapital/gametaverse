TARGET:=twitterspy
TAG=v0.1.2
DIST_DIR=../dist_$(TARGET)_$(TAG)

SRC_FILES := $(shell find ../ -type  f | grep "\.go"   )
#SRC_FILES += main.go flag.go config_xml.go

.DEFAULT:all win mac

all:$(TARGET) 
	@echo "Done All." 


gencodec: 
	cd ../../; go generate

$(TARGET):$(SRC_FILES)
	go build -o $(TARGET)
	@echo "Done building."

win:$(SRC_FILES)
	GOOS=windows GOARCH=amd64 go build -o $(TARGET).exe
	@echo "Done building windows."

mac:$(SRC_FILES)
	GOOS=darwin GOARCH=amd64 go build -o $(TARGET).bin
	@echo "Done building mac."

dist:$(TARGET)
	rm -rf ../dist_*
	mkdir $(DIST_DIR)
	mkdir $(DIST_DIR)/linux
	cp $(TARGET) $(DIST_DIR)/linux
	cp -rf ../scripts $(DIST_DIR)/linux

.PHONY: clean 
clean:
	rm -rf $(TARGET) $(TARGET).exe $(TARGET).bin
	go clean -cache