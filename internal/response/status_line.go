package response

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func GetReasonPhrase(statusCode StatusCode) string {
	reasonPhrase := ""
	switch statusCode {
	case 200:
		reasonPhrase = "OK"
	case 400:
		reasonPhrase = "Bad Request"
	case 500:
		reasonPhrase = "Internal Server Error"
	}

	return reasonPhrase
}
