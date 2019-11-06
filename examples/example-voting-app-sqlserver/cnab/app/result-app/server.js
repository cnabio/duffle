console.log("starting")
var express = require('express'),
    mssql = require('tedious'),
    cookieParser = require('cookie-parser'),
    bodyParser = require('body-parser'),
    methodOverride = require('method-override'),
    app = express(),
    server = require('http').Server(app),
    io = require('socket.io')(server);

io.set('transports', ['polling']);



var port = process.env.PORT || 4000;

io.sockets.on('connection', function (socket) {
  socket.emit('message', { text : 'Welcome!' });
  socket.on('subscribe', function (data) {
    socket.join(data.channel);
  });
});

//Use setInterval to retry the connection the SQL Server until it connects to allow time for the SQL Server container to come up
var connectionAttempts = 1;
var connected = false;
console.log("Attempting connection...");
setInterval(function () {
  if(!connected) {
    var Connection = mssql.Connection;
    var config = {
        userName: 'sa'
        ,password: 'SuperSecretPassword1234'
        ,server: 'db'
        ,options: {database:'Votes'}
        /* Uncomment if you need to debug connections/queries
        ,debug:
          {
          packet: false,
          data: false,
          payload: false,
          token: false,
          log: false
          }*/
        };
    var connection = new Connection(config);
    console.log("Connection attempt: " + connectionAttempts);
    connectionAttempts++;
    /* Uncomment if you need to debug connections/queries
    connection.on('infoMessage', infoError);
    connection.on('errorMessage', infoError);
    connection.on('debug', debug);
    */
    connection.on('connect', function(err) {
      if (err) {
        console.log(err);
        connected = false;
      } else {
        connected = true;
        console.log("Connected");
        setInterval(function() { getVotes(connection); }, 1000);
      }
    });
  }
}
, 10000);

/* Uncomment if you need to debug connections/queries
function infoError(info) {
  var dd = info;
  console.log('infoError=> ' + info);
}

function debug(message) {
  var dd = message;
  console.log('debug=> ' + message);
}
*/

function getVotes(connection) {
  console.log("Getting Votes...");
  var Request = require('tedious').Request;
  request = new Request("SELECT (SELECT COUNT(*) FROM votes WHERE vote = 'a') AS 'a'\
                        , (SELECT COUNT(*) FROM votes WHERE vote = 'b') AS 'b' \
                        FOR JSON PATH, WITHOUT_ARRAY_WRAPPER;", function(err, rowCount) {
        if (err) {
            console.log(err);
        }
  });
  
  request.on('row', function(columns) { 
            columns.forEach(function(column) {  
              if (column.value === null) {  
                console.log('NULL');
              } else {  
                io.sockets.emit("scores",column.value);
              }  
            });
  });

  connection.execSql(request);  
}

app.use(cookieParser());
app.use(bodyParser());
app.use(methodOverride('X-HTTP-Method-Override'));
app.use(function(req, res, next) {
  res.header("Access-Control-Allow-Origin", "*");
  res.header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept");
  res.header("Access-Control-Allow-Methods", "PUT, GET, POST, DELETE, OPTIONS");
  next();
});

app.use(express.static(__dirname + '/views'));

app.get('/', function (req, res) {
  res.sendFile(path.resolve(__dirname + '/views/index.html'));
});

server.listen(port, function () {
  var port = server.address().port;
  console.log('App running on port ' + port);
});
