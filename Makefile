build:
	go build -o zap
cp:
	cp zap ~/.local/bin/
	
install: build cp