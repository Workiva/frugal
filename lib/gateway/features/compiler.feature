Feature: Frugal HTTP+JSON Gateway
  In order to release a Frugal RPC service as a third-party API
  As a Workiva software developer
  I want to compile my service's IDL file into a runnable HTTP proxy.

  Background:
    Given valid IDL with body
      """
      namespace * v1.music

      struct Album {
        1: string artist
        2: string title
        3: string ASIN
      }

      struct LookupAlbumRequest {
        1: string ASIN (http.jsonProperty="asin")
      }

      struct BuyAlbumRequest {
        1: string artist (http.jsonProperty="artist")
        2: string title (http.jsonProperty="title")
      }
      """

  @working
  Scenario: Basic get request
    Given a service method with annotations like
      """
      service Store {
        Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album" with payload 
      """
      {"asin": "abcdefg"}
      """
    Then the response should be
      """
      {
        "artist": "Coeur de Pirates",
        "title": "Comme des enfants",
        "ASIN": "abcdefg"
      }
      """

  @working
  Scenario: Basic post request
    Given a service method with annotations like
      """
      service Store {
        Album buyAlbum(1: BuyAlbumRequest request) (http.method="post", http.pathTemplate="/v1/store/album")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album" with payload 
      """
      {"asin": "abcdefg"}
      """
    Then the response should be
      """
      {
        "artist": "Coeur de Pirates",
        "title": "Comme des enfants",
        "ASIN": "abcdefg"
      }
      """


  @working
  Scenario: Path parameter get request
    Given a service method with annotations like
      """
      service Store {
        Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album/{asin}")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album/agcdefg" with payload 
      """
      """
    Then the response should be
      """
      {
        "artist": "Coeur de Pirates",
        "title": "Comme des enfants",
        "ASIN": "abcdefg"
      }
      """

  @working
  Scenario: Query parameter get request
    Given a service method with annotations like
      """
      service Store {
      Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album/", http.query="asin")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album/?asin=abcdefg" with payload 
      """
      """
    Then the response should be
      """
      {
        "artist": "Coeur de Pirates",
        "title": "Comme des enfants",
        "ASIN": "abcdefg"
      }
      """

  @working
  Scenario: Invalid JSON syntax
    Given a service method with annotations like
      """
      service Store {
        Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album" with payload 
      """
      {<invalid-json"|}
      """
    Then the response should be
      """
      { "message": "Problems parsing JSON" }
      """

  @working
  Scenario: Invalid JSON body
    Given a service method with annotations like
      """
      service Store {
        Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album" with payload 
      """
      {"name": "Kevin"}
      """
    Then the response should be
      """
      { "message": "Invalid JSON data" }
      """

  @working
  Scenario: Invalid JSON values
    Given a service method with annotations like
      """
      service Store {
        Album lookupAlbum(1: LookupAlbumRequest request) (http.pathTemplate="/v1/store/album")
      } 
      """
    And  the compiler generates a Frugal processor
    And  a Frugal server is running
    And  the compiler generates HTTP proxy handlers
    And  a proxy server is running
    When a "GET" request is made to "/v1/store/album" with payload 
      """
      {"asin": 7}
      """
    Then the response should be
      """
      {
        "message": "Validation failed",
        "errors": [
          {
            "resource": "album",
            "field": "asin",
            "code": "invalid"
          }
        ]
      }
      """
