# Тестовое задание для компании Sedmax

Пример команды для тестирования:

```
modbus_client --debug -mtcp -t0x10 -r10 -p1502 127.0.0.1 0x01 0x02 0x03 0x04 0x05
modbus_client --debug -mtcp -t0x03 -r10 -p1502 127.0.0.1 -с5
```

Для запуска Unit тестов: `go test` из корня проекта