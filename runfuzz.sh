clear
# Cant wildcardrun them in a row
go test -v -fuzztime 5m -parallel 4 -fuzz=FuzzAccountStore
go test -v -fuzztime 5m -parallel 4 -fuzz=FuzzWindowsErrorCodes
go test -v -fuzztime 5m -parallel 4 -fuzz=FuzzSanitiseParseCommandWithArguments
