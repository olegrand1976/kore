import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';

import '../api/api_client.dart';
import 'auth_repository.dart';

/// OIDC PKCE via Kore API broker — opens IdP in system browser, catches kore:// callback.
class OidcService {
  OidcService({
    required ApiClient apiClient,
    required AuthRepository authRepository,
    this.redirectUri = 'kore://callback',
  })  : _api = apiClient,
        _auth = authRepository;

  final ApiClient _api;
  final AuthRepository _auth;
  final String redirectUri;

  Future<String> buildAuthorizeUrl({
    required String tenantId,
    required PkceChallenge pkce,
  }) async {
    final envelope = await _api.get(
      '/auth/oidc/authorize',
      query: {
        'tenant': tenantId,
        'redirect_uri': redirectUri,
        'code_challenge': pkce.challenge,
        'state': pkce.state,
      },
    );
    return envelope['data']['authorizeUrl'] as String;
  }

  Future<AuthSession> signInWithSso({required String tenantId}) async {
    final pkce = _auth.generatePkce();
    await _auth.saveOidcState(pkce.state, pkce.verifier);
    final authorizeUrl = await buildAuthorizeUrl(tenantId: tenantId, pkce: pkce);

    final callback = await FlutterWebAuth2.authenticate(
      url: authorizeUrl,
      callbackUrlScheme: Uri.parse(redirectUri).scheme,
    );

    final uri = Uri.parse(callback);
    final code = uri.queryParameters['code'];
    final state = uri.queryParameters['state'];
    if (code == null || code.isEmpty || state == null || state.isEmpty) {
      throw StateError('OIDC callback missing code or state');
    }
    return completeCallback(tenantId: tenantId, code: code, state: state);
  }

  Future<AuthSession> completeCallback({
    required String tenantId,
    required String code,
    required String state,
  }) async {
    final verifier = await _auth.consumeOidcVerifier(state);
    final session = await _api.postData(
      '/auth/oidc/callback',
      body: {
        'tenantId': tenantId,
        'code': code,
        'redirectUri': redirectUri,
        'codeVerifier': verifier,
        'state': state,
      },
      parser: (json) => AuthSession.fromJson(json as Map<String, dynamic>),
    );
    await _auth.persistSession(session);
    return session;
  }
}
