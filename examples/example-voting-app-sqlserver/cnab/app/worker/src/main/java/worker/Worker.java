package worker;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.exceptions.JedisConnectionException;
import java.sql.*;
import com.microsoft.sqlserver.jdbc.*;
import org.json.JSONObject;

class Worker {
  public static void main(String[] args) {
    try {
      Jedis redis = connectToRedis("redis");
      Connection dbConn = connectToDB("db");

      System.err.println("Watching vote queue");

      while (true) {
        String voteJSON = redis.blpop(0, "votes").get(1);
        JSONObject voteData = new JSONObject(voteJSON);
        String voterID = voteData.getString("voter_id");
        String vote = voteData.getString("vote");

        System.err.printf("Processing vote for '%s' by '%s'\n", vote, voterID);
        updateVote(dbConn, voterID, vote);
      }
    } catch (SQLException e) {
      e.printStackTrace();
      System.exit(1);
    }
  }

  static void updateVote(Connection dbConn, String voterID, String vote) throws SQLException {
    PreparedStatement insert = dbConn.prepareStatement(
      "INSERT INTO votes (id, vote) VALUES (?, ?)");
    insert.setString(1, voterID);
    insert.setString(2, vote);

    try {
      insert.executeUpdate();
    } catch (SQLException e) {
      PreparedStatement update = dbConn.prepareStatement(
        "UPDATE votes SET vote = ? WHERE id = ?");
      update.setString(1, vote);
      update.setString(2, voterID);
      update.executeUpdate();
    }
  }

  static Jedis connectToRedis(String host) {
    Jedis conn = new Jedis(host);

    while (true) {
      try {
        conn.keys("*");
        break;
      } catch (JedisConnectionException e) {
        System.err.println("Failed to connect to redis - retrying");
        sleep(1000);
      }
    }

    System.err.println("Connected to redis");
    return conn;
  }

  static Connection connectToDB(String host) throws SQLException {
    Connection conn = null;

    try {

      Class.forName("com.microsoft.sqlserver.jdbc.SQLServerDriver");
      String url = "jdbc:sqlserver://" + host + ";database=master;user=sa;password=SuperSecretPassword1234";

      while (conn == null) {
        try {
          conn = DriverManager.getConnection(url);
        } catch (SQLException e) {
          System.err.println(e);
          sleep(1000);
        }
      }
      
      String dropDB =  
            "IF EXISTS(SELECT * from sys.databases WHERE name='Votes')"  
          + " BEGIN"  
          + " DROP DATABASE Votes;"
          + " END";
          
      PreparedStatement stDropDB = conn.prepareStatement(dropDB);
      stDropDB.executeUpdate();
      
      String createDB = "CREATE DATABASE Votes;";
      PreparedStatement stCreateDB = conn.prepareStatement(createDB);
      stCreateDB.executeUpdate();
      
      String createTable = "USE Votes; CREATE TABLE votes (id VARCHAR(255) NOT NULL PRIMARY KEY, vote VARCHAR(255) NOT NULL);";
      PreparedStatement stCreateTable = conn.prepareStatement(createTable);
      stCreateTable.executeUpdate();
      
    } catch (ClassNotFoundException e) {
      e.printStackTrace();
      System.exit(1);
    }

    return conn;
  }

  static void sleep(long duration) {
    try {
      Thread.sleep(duration);
    } catch (InterruptedException e) {
      System.exit(1);
    }
  }
}
