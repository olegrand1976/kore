import 'dart:convert';

import 'package:http/http.dart' as http;

import '../auth/auth_repository.dart';
import 'api_exceptions.dart';

typedef TokenProvider = Future<String?> Function();

class ApiClient {
  ApiClient({
    required this.baseUrl,
    required AuthRepository authRepository,
    http.Client? httpClient,
  })  : _auth = authRepository,
        _http = httpClient ?? http.Client();

  final String baseUrl;
  final AuthRepository _auth;
  final http.Client _http;

  Uri _uri(String path, [Map<String, String>? query]) {
    final normalized = path.startsWith('/') ? path : '/$path';
    return Uri.parse('$baseUrl$normalized').replace(queryParameters: query);
  }

  Future<Map<String, dynamic>> get(
    String path, {
    Map<String, String>? query,
  }) async {
    return _send(() async {
      final response = await _http.get(
        _uri(path, query),
        headers: await _headers(),
      );
      return _handleResponse(response);
    });
  }

  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    return _send(() async {
      final response = await _http.post(
        _uri(path),
        headers: await _headers(),
        body: jsonEncode(body ?? {}),
      );
      return _handleResponse(response);
    });
  }

  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    return _send(() async {
      final response = await _http.put(
        _uri(path),
        headers: await _headers(),
        body: jsonEncode(body ?? {}),
      );
      return _handleResponse(response);
    });
  }

  Future<T> getData<T>(
    String path, {
    Map<String, String>? query,
    required T Function(dynamic json) parser,
  }) async {
    final envelope = await get(path, query: query);
    return parser(envelope['data']);
  }

  Future<T> postData<T>(
    String path, {
    Map<String, dynamic>? body,
    required T Function(dynamic json) parser,
  }) async {
    final envelope = await post(path, body: body);
    return parser(envelope['data']);
  }

  Future<T> putData<T>(
    String path, {
    Map<String, dynamic>? body,
    required T Function(dynamic json) parser,
  }) async {
    final envelope = await put(path, body: body);
    return parser(envelope['data']);
  }

  Future<Map<String, String>> _headers() async {
    final token = await _auth.accessToken;
    return {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      if (token != null && token.isNotEmpty) 'Authorization': 'Bearer $token',
    };
  }

  Future<Map<String, dynamic>> _send(
    Future<Map<String, dynamic>> Function() request,
  ) async {
    try {
      return await request();
    } on TokenExpiredException {
      await _auth.refreshTokens();
      return request();
    }
  }

  Map<String, dynamic> _handleResponse(http.Response response) {
    final body = response.body.isEmpty
        ? <String, dynamic>{}
        : jsonDecode(response.body) as Map<String, dynamic>;

    if (response.statusCode == 401) {
      final error = body['error'] as Map<String, dynamic>?;
      final code = error?['code'] as String? ?? '';
      if (code == 'TOKEN_EXPIRED') {
        throw TokenExpiredException();
      }
      throw UnauthorizedException(
        error?['message'] as String? ?? 'unauthorized',
      );
    }

    if (response.statusCode >= 400) {
      final error = body['error'] as Map<String, dynamic>?;
      throw ApiException(
        response.statusCode,
        error?['code'] as String? ?? 'ERROR',
        error?['message'] as String? ?? response.reasonPhrase ?? 'error',
      );
    }

    return body;
  }

  void dispose() => _http.close();
}
