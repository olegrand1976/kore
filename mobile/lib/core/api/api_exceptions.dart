class ApiException implements Exception {
  ApiException(this.statusCode, this.code, this.message);

  final int statusCode;
  final String code;
  final String message;

  @override
  String toString() => 'ApiException($statusCode, $code): $message';
}

class UnauthorizedException extends ApiException {
  UnauthorizedException(String message)
      : super(401, 'UNAUTHORIZED', message);
}

class TokenExpiredException extends ApiException {
  TokenExpiredException()
      : super(401, 'TOKEN_EXPIRED', 'access token expired');
}
