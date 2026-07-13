import 'dart:convert';
import 'dart:math';

import 'package:crypto/crypto.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:http/http.dart' as http;

import '../api/api_exceptions.dart';

class AuthSession {
  const AuthSession({
    required this.accessToken,
    required this.refreshToken,
    this.userId,
    this.tenantId,
    this.profile,
  });

  final String accessToken;
  final String refreshToken;
  final String? userId;
  final String? tenantId;
  final String? profile;

  factory AuthSession.fromJson(Map<String, dynamic> json) {
    return AuthSession(
      accessToken: _pickString(json, ['accessToken', 'AccessToken']) ?? '',
      refreshToken: _pickString(json, ['refreshToken', 'RefreshToken']) ?? '',
      userId: _pickString(json, ['userId', 'UserID']),
      tenantId: _pickString(json, ['tenantId', 'TenantID']),
      profile: _pickString(json, ['profile', 'Profile']),
    );
  }

  static String? _pickString(Map<String, dynamic> json, List<String> keys) {
    for (final key in keys) {
      final value = json[key];
      if (value != null) return value.toString();
    }
    return null;
  }
}

class AuthRepository {
  AuthRepository({
    required this.baseUrl,
    FlutterSecureStorage? storage,
    http.Client? httpClient,
  })  : _storage = storage ?? const FlutterSecureStorage(),
        _http = httpClient ?? http.Client();

  final String baseUrl;
  final FlutterSecureStorage _storage;
  final http.Client _http;

  static const _accessKey = 'kore_access_token';
  static const _refreshKey = 'kore_refresh_token';
  static const _profileKey = 'kore_profile';
  static const _oidcStatePrefix = 'kore_oidc_state_';

  AuthSession? _cached;

  Future<String?> get accessToken async {
    _cached ??= await loadSession();
    return _cached?.accessToken;
  }

  Future<AuthSession?> loadSession() async {
    final access = await _storage.read(key: _accessKey);
    if (access == null || access.isEmpty) return null;
    _cached = AuthSession(
      accessToken: access,
      refreshToken: await _storage.read(key: _refreshKey) ?? '',
      profile: await _storage.read(key: _profileKey),
    );
    return _cached;
  }

  Future<bool> get isAuthenticated async {
    final token = await accessToken;
    return token != null && token.isNotEmpty;
  }

  Future<AuthSession> loginWithPassword({
    required String login,
    required String password,
  }) async {
    final response = await _http.post(
      Uri.parse('$baseUrl/auth/login'),
      headers: const {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode({'login': login, 'password': password}),
    );
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    if (response.statusCode >= 400) {
      final error = body['error'] as Map<String, dynamic>?;
      throw ApiException(
        response.statusCode,
        error?['code'] as String? ?? 'ERROR',
        error?['message'] as String? ?? 'login failed',
      );
    }
    final session = AuthSession.fromJson(body['data'] as Map<String, dynamic>);
    await persistSession(session);
    return session;
  }

  Future<void> persistSession(AuthSession session) async {
    _cached = session;
    await _storage.write(key: _accessKey, value: session.accessToken);
    await _storage.write(key: _refreshKey, value: session.refreshToken);
    if (session.profile != null) {
      await _storage.write(key: _profileKey, value: session.profile);
    }
  }

  Future<void> refreshTokens() async {
    final refresh = await _storage.read(key: _refreshKey);
    if (refresh == null || refresh.isEmpty) {
      throw UnauthorizedException('no refresh token');
    }
    final response = await _http.post(
      Uri.parse('$baseUrl/auth/refresh'),
      headers: const {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: jsonEncode({'refreshToken': refresh}),
    );
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    if (response.statusCode >= 400) {
      await logout();
      final error = body['error'] as Map<String, dynamic>?;
      throw ApiException(
        response.statusCode,
        error?['code'] as String? ?? 'ERROR',
        error?['message'] as String? ?? 'refresh failed',
      );
    }
    final data = body['data'] as Map<String, dynamic>;
    final session = AuthSession(
      accessToken: _pick(data, ['accessToken', 'AccessToken']) ?? '',
      refreshToken: _pick(data, ['refreshToken', 'RefreshToken']) ?? refresh,
      profile: _cached?.profile,
    );
    await persistSession(session);
  }

  Future<void> logout() async {
    final refresh = await _storage.read(key: _refreshKey);
    if (refresh != null && refresh.isNotEmpty) {
      try {
        await _http.post(
          Uri.parse('$baseUrl/auth/logout'),
          headers: const {'Content-Type': 'application/json'},
          body: jsonEncode({'refreshToken': refresh}),
        );
      } catch (_) {
        // Best-effort server invalidation.
      }
    }
    _cached = null;
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
    await _storage.delete(key: _profileKey);
  }

  Future<void> saveOidcState(String state, String verifier) async {
    await _storage.write(key: '$_oidcStatePrefix$state', value: verifier);
  }

  Future<String> consumeOidcVerifier(String state) async {
    final key = '$_oidcStatePrefix$state';
    final verifier = await _storage.read(key: key);
    await _storage.delete(key: key);
    if (verifier == null || verifier.isEmpty) {
      throw StateError('OIDC state expired or invalid');
    }
    return verifier;
  }

  PkceChallenge generatePkce() {
    final random = Random.secure();
    final verifierBytes = List<int>.generate(32, (_) => random.nextInt(256));
    final verifier = base64Url.encode(verifierBytes).replaceAll('=', '');
    final challenge = base64Url
        .encode(sha256.convert(utf8.encode(verifier)).bytes)
        .replaceAll('=', '');
    final state = base64Url
        .encode(List<int>.generate(16, (_) => random.nextInt(256)))
        .replaceAll('=', '');
    return PkceChallenge(
      verifier: verifier,
      challenge: challenge,
      state: state,
    );
  }

  String? _pick(Map<String, dynamic> json, List<String> keys) {
    for (final key in keys) {
      final value = json[key];
      if (value != null) return value.toString();
    }
    return null;
  }

  void dispose() => _http.close();
}

class PkceChallenge {
  const PkceChallenge({
    required this.verifier,
    required this.challenge,
    required this.state,
  });

  final String verifier;
  final String challenge;
  final String state;
}
