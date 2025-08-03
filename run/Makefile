all:
	@mkdir -p build
	@bear --output build/compile_commands.json -- make o2.so

o2.so: run.c
	$(CC) $(CFLAGS) -shared -o o2.so -I./include $^
