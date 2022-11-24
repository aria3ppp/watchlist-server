package response

//go:generate go-enum --file=$GOFILE --marshal

/*
ENUM(
OK
NotFound
InvalidURLParameter
InvalidRequest
EmailAlreadyUsed
EmailNotFound
IncorrectPassword
SameNewPassword
TokenInvalid
TokenMissingOrMalformed
InternalServerError
)
*/
type Status int
