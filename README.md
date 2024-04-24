# dam

Demo:

```
// mockdam
cd mockdam && go run .

// dam
go run . -p ./mockdam/ttyIN

// visit http://127.0.0.1:15161 and run these commands:
// OpenAll
// CloseAll
// OpenClose 1
// CloseOpen 1
// Close 1
// Open 1
// (currently only #01 is OK)
```

If you are using USB-Serial cable, you may need to download the driver.

For example, if you're using ugreen cable and macOS (including Apple Silicon),
you can download the driver here: <https://www.lulian.cn/download/17.html>.
