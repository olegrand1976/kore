import 'package:flutter/foundation.dart';

import '../api/api_client.dart';

/// Enregistrement des tokens push auprès du backend Kore.
/// FCM natif : ajouter firebase_messaging + google-services.json / GoogleService-Info.plist.
class PushService {
  PushService({required ApiClient apiClient}) : _api = apiClient;

  final ApiClient _api;
  String? _lastToken;

  Future<void> registerToken({
    required String platform,
    required String token,
  }) async {
    if (token.isEmpty || token == _lastToken) {
      return;
    }
    await _api.post(
      '/devices/register',
      body: {'platform': platform, 'token': token},
    );
    _lastToken = token;
    if (kDebugMode) {
      debugPrint('push: token registered ($platform)');
    }
  }

  Future<void> unregisterToken(String token) async {
    if (token.isEmpty) {
      return;
    }
    await _api.delete(
      '/devices/register',
      body: {'token': token},
    );
    if (_lastToken == token) {
      _lastToken = null;
    }
  }

  /// Point d'entrée FCM — à brancher sur firebase_messaging quand la config native est prête.
  Future<void> initMessaging() async {
    if (kDebugMode) {
      debugPrint('push: FCM non configuré — enregistrement manuel via registerToken');
    }
  }
}
