build:
	go build -o lif
cp:
	cp lif ~/.local/bin/
	
install: build cp