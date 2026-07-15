import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';

import '../api/api_client.dart';

/// Enregistrement des tokens push auprès du backend Kore.
/// FCM natif : ajouter `firebase_core` + `firebase_messaging` et les fichiers Google Services.
class PushService {
  PushService({required ApiClient apiClient}) : _api = apiClient;

  final ApiClient _api;
  String? _lastToken;

  String get _platform {
    if (kIsWeb) {
      return 'web';
    }
    if (Platform.isIOS) {
      return 'ios';
    }
    if (Platform.isAndroid) {
      return 'android';
    }
    return 'web';
  }

  Future<void> registerToken({
    String? platform,
    required String token,
  }) async {
    if (token.isEmpty || token == _lastToken) {
      return;
    }
    await _api.post(
      '/devices/register',
      body: {'platform': platform ?? _platform, 'token': token},
    );
    _lastToken = token;
    if (kDebugMode) {
      debugPrint('push: token registered (${platform ?? _platform})');
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

  /// Initialise FCM quand Firebase est configuré ; sinon no-op silencieux.
  Future<void> initMessaging() async {
    if (kDebugMode) {
      debugPrint(
        'push: FCM natif non lié — exécuter flutterfire configure puis brancher firebase_messaging',
      );
    }
  }

  Future<void> syncAfterLogin() async {
    await initMessaging();
  }
}
