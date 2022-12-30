package response

//go:generate go-enum --file=$GOFILE --marshal

/*
ENUM(
OK
NotFound
InvalidURLParameter
InvalidRequest
UsedEmail
IncorrectPassword
SamePassword
InvalidToken
MissingToken
MissingFile
UnsupportedMediaType
InternalServerError
)
*/
type Status int
