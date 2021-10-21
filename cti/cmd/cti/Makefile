TARGET:=cti

SRC_FILES := $(shell find ../.. -type  f | grep "\.go"   )
#SRC_FILES += main.go flag.go config_xml.go

.DEFAULT:all

all:$(TARGET) 
	@echo "Done All." 


gencodec: 
	cd ../../; go generate

$(TARGET):$(SRC_FILES)
	go build -o $(TARGET)
	@echo "Done building."


.PHONY: clean 
clean:
	rm -rf $(TARGET)
	go clean -cache
