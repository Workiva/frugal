import 'dart:async';
import 'dart:html';

import 'package:logging/logging.dart';
import 'package:thrift/thrift.dart';
import 'package:v1_music/v1_music.dart' as music;
import 'package:frugal/frugal.dart' as frugal;

frugal.FSubscription sub;

void main() {
  Logger.root.level = Level.FINEST;
  Logger.root.onRecord.listen((LogRecord r) {
    window.console.log('${r.loggerName}(${r.level}): ${r.message}');
  });
  new EventUI(querySelector('#output')).start();
}

class EventUI {
  final DivElement output;

  EventUI(this.output);

  music.AlbumWinnersPublisher _albumWinnersPublisher;
  music.AlbumWinnersSubscriber _albumWinnersSubscriber;

  music.FStoreClient _fStoreClient;

  frugal.Middleware loggingMiddleware() {
    return (frugal.InvocationHandler next) {
      return (String serviceName, String methodName, List<Object> args) {
        print("==== CALLING $serviceName.$methodName ====");
        var ret = next(serviceName, methodName, args);
        print("==== CALLED  $serviceName.$methodName ====");
        return ret;
      };
    };
  }

  void start() {
    _buildInterface();
    _initConnection();
  }

  void _initConnection() { }

  void _buildInterface() {
    output.children.forEach((e) {
      e.remove();
    });

    _buildPublishComponent();
    _buildSubscribeComponent();
    _buildRequestComponent();
  }

  void _buildPublishComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Publish Album Winner");
    InputElement asin = new InputElement()
      ..id = "asin"
      ..type = "string";
    output.append(asin);
    InputElement duration = new InputElement()
      ..id = "duration"
      ..type = "number";
    output.append(duration);
    ButtonElement publishButton = new ButtonElement()
      ..text = "Publish"
      ..onClick.listen(_onPublishClick);
    output.append(publishButton);
  }

  void _onPublishClick(MouseEvent e) {
    InputElement asin = querySelector("#asin");
    InputElement duration = querySelector("#duration");
    var album = new music.Album();
    album.ASIN = asin.value;
    album.duration = int.parse(duration.value);
    frugal.FContext ctx = new frugal.FContext(correlationId: 'an-id');
    _albumWinnersPublisher.publishWinner(ctx, album);
  }

  void _buildSubscribeComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Subscribe To Album Winner Event");
    ButtonElement subscribeButton = new ButtonElement()
      ..text = "Subscribe"
      ..onClick.listen(_onSubscribeClick);
    output.append(subscribeButton);
    ButtonElement unsubscribeButton = new ButtonElement()
      ..text = "Unsubscribe"
      ..onClick.listen(_onUnsubscribeClick);
    output.append(unsubscribeButton);
  }

  Future _onSubscribeClick(MouseEvent e) async {
    if (sub == null ){
      sub = await _albumWinnersSubscriber.subscribeWinner(onEvent);
    }
  }

  Future _onUnsubscribeClick(MouseEvent e) async {
    if (sub != null ){
      await sub.unsubscribe();
      sub = null;
    }
  }

  void _buildRequestComponent() {
    output.append(new HeadingElement.h3()
      ..text = "Music Service");

    ButtonElement buyAlbumButton = new ButtonElement()
      ..text  = "Buy Album"
      ..onClick.listen(_onBuyAlbumClick);
    output.append(buyAlbumButton);
  }

  void _onBuyAlbumClick(MouseEvent e) {
    if (_fStoreClient == null) {
      window.alert("Not connected to server");
    }
    var ctx = new frugal.FContext(correlationId:"corr-12345");
    _fStoreClient.buyAlbum(ctx, "My-ASIN", "Account-12345").catchError( (e) {
      window.alert("Ping errored! ${e.toString()}");
    });
  }

  void onEvent(frugal.FContext ctx, music.Album m) {
    window.alert(ctx.correlationId().toString() + ' : ' + m.toString());
  }
}

