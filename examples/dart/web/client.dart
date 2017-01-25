import 'dart:async';
import 'dart:html';

import 'package:logging/logging.dart';
import 'package:v1_music/v1_music.dart' as music;
import 'package:thrift/thrift.dart' as thrift;
import 'package:frugal/frugal.dart' as frugal;
import 'package:w_transport/w_transport.dart' as wt;
import 'package:w_transport/w_transport_browser.dart'
    show configureWTransportForBrowser;

frugal.FSubscription sub;

void main() {
  configureWTransportForBrowser();
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

  _initConnection() async {
    var uri = Uri.parse("http://localhost:9090/frugal");
    var transport = new frugal.FHttpTransport(new wt.Client(), uri);
    await transport.open();

    // Wire up FServiceProvider
    var tBinaryProtocolFactory = new thrift.TBinaryProtocolFactory();
    var protocolFactory = new frugal.FProtocolFactory(tBinaryProtocolFactory);
    var provider = new frugal.FServiceProvider(transport, protocolFactory);

    _fStoreClient = new music.FStoreClient(provider);
  }

  void _buildInterface() {
    output.children.forEach((e) {
      e.remove();
    });

    _buildPublishComponent();
    _buildSubscribeComponent();
    _buildRequestComponent();
  }

  void _buildPublishComponent() {
    output.append(new HeadingElement.h3()..text = "Publish Album Winner");
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
    album.aSIN = asin.value;
    album.duration = int.parse(duration.value);
    frugal.FContext ctx = new frugal.FContext(correlationId: 'an-id');
    _albumWinnersPublisher.publishWinner(ctx, album);
  }

  void _buildSubscribeComponent() {
    output.append(
        new HeadingElement.h3()..text = "Subscribe To Album Winner Event");
    output.append(new HeadingElement.h4()
      ..text = "(SDK Only - Not Implemented in frugal)");
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
    if (sub == null) {
      sub = await _albumWinnersSubscriber.subscribeWinner(onEvent);
    }
  }

  Future _onUnsubscribeClick(MouseEvent e) async {
    if (sub != null) {
      await sub.unsubscribe();
      sub = null;
    }
  }

  void _buildRequestComponent() {
    output.append(new HeadingElement.h3()..text = "Music Service");

    ButtonElement buyAlbumButton = new ButtonElement()
      ..text = "Buy Album"
      ..onClick.listen(_onBuyAlbumClick);
    output.append(buyAlbumButton);
  }

  Future _onBuyAlbumClick(MouseEvent e) async {
    if (_fStoreClient == null) {
      window.alert("Not connected to server");
    }
    var ctx = new frugal.FContext(correlationId: "corr-12345");
    var album = await _fStoreClient
        .buyAlbum(ctx, "My-ASIN", "Account-12345")
        .catchError((e) {
      window.alert("Ping errored! ${e.toString()}");
    });

    window.alert("Bought album: $album");
  }

  void onEvent(frugal.FContext ctx, music.Album m) {
    window.alert(ctx.correlationId.toString() + ' : ' + m.toString());
  }
}
