import '../api/api_client.dart';
import 'auth_repository.dart';

/// OIDC PKCE stub — builds authorize URL via Kore API, exchanges code on callback.
///
/// Production: wire [FlutterAppAuth] or custom tabs to open [authorizeUrl]
/// and invoke [completeCallback] with the redirect query params.
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

  Future<String> buildAuthorizeUrl({required String tenantId}) async {
    final pkce = _auth.generatePkce();
    await _auth.saveOidcState(pkce.state, pkce.verifier);

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

  /// Stub: opens SSO in system browser — integrate flutter_appauth here.
  Future<void> signInWithSso({required String tenantId}) async {
    final url = await buildAuthorizeUrl(tenantId: tenantId);
    throw UnimplementedError(
      'OIDC browser redirect not wired yet. Open: $url',
    );
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
