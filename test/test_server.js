
var request = require('request');
var assert = require('assert');
var server = require('../src/lib/server');
var Jwt = require('../src/lib/jwt');
var speakeasy = require('speakeasy');
var sinon = require('sinon');

describe('test the server', function() {
  var jwt = new Jwt('jwt_secret');
  var ldap_client = {
    bind: sinon.stub()
  };

  before(function() {
    var config = {
      port: 8080,
      totp_secret: 'totp_secret',
      ldap_url: 'ldap://127.0.0.1:389',
      ldap_users_dn: 'ou=users,dc=example,dc=com',
      jwt_secret: 'jwt_secret',
      jwt_expiration_time: '1h'
    };

    // ldap_client.bind.yields(undefined);
    ldap_client.bind.withArgs('cn=test_ok,ou=users,dc=example,dc=com', 
                              'password').yields(undefined);
    ldap_client.bind.withArgs('cn=test_nok,ou=users,dc=example,dc=com', 
                              'password').yields('error');
    server.run(config, ldap_client);
  });


  describe('test GET /login', function() {
    test_login()
  });

  describe('test GET /logout', function() {
    test_logout()
  });

  describe('test GET /_auth', function() {
    test_get_auth(jwt);
  });
 
  describe('test POST /_auth', function() {
    test_post_auth(jwt);
  });
});

function test_login() {
  it('should serve the login page', function(done) {
    request.get('http://localhost:8080/login')
    .on('response', function(response) {
      assert.equal(response.statusCode, 200);
      done();
    }) 
  });
}

function test_logout() {
  it('should logout and redirect to /', function(done) {
    request.get('http://localhost:8080/logout')
    .on('response', function(response) {
      assert.equal(response.req.path, '/');
      done();
    }) 
  });
}

function test_get_auth(jwt) {
  it('should return status code 401 when user is not authenticated', function(done) {
    request.get('http://localhost:8080/_auth')
    .on('response', function(response) {
      assert.equal(response.statusCode, 401);
      done();
    }) 
  });

  it('should return status code 204 when user is authenticated', function(done) {
    var j = request.jar();
    var r = request.defaults({jar: j});
    var token = jwt.sign({ user: 'test' }, '1h');
    var cookie = r.cookie('access_token=' + token);
    j.setCookie(cookie, 'http://localhost:8080/_auth');

    r.get('http://localhost:8080/_auth')
    .on('response', function(response) {
      assert.equal(response.statusCode, 204);
      done();
    }) 
  });
}

function test_post_auth() {
  it('should return the JWT token when authentication is successful', function(done) {
    var clock = sinon.useFakeTimers();
    var real_token = speakeasy.totp({
      secret: 'totp_secret',
      encoding: 'base32'
    });
    var expectedJwt = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoidGVzdF9vayIsImlhdCI6MCwiZXhwIjozNjAwfQ.ihvaljGjO5h3iSO_h3PkNNSCYeePyB8Hr5lfVZZYyrQ';

    request.post('http://localhost:8080/_auth', { 
      form: {
        username: 'test_ok',
        password: 'password',
        token: real_token
      }
    }, 
    function (error, response, body) {
      if (!error && response.statusCode == 200) {
        assert.equal(body, expectedJwt);
        clock.restore();
        done();
      }
    });
  });

  it('should return invalid authentication status code', function(done) {
    var clock = sinon.useFakeTimers();
    var real_token = speakeasy.totp({
      secret: 'totp_secret',
      encoding: 'base32'
    });
    var data = {
      form: {
        username: 'test_nok',
        password: 'password',
        token: real_token
      }
    }

    request.post('http://localhost:8080/_auth', data, function (error, response, body) {
      if(response.statusCode == 401) {
        clock.restore();
        done();
      }
    });
  });
}