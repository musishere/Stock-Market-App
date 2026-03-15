import React, { useEffect, useState } from 'react';
import {
  SafeAreaView,
  View,
  Text,
  FlatList,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { w3cwebsocket as W3CWebSocket } from 'websocket';

const WS_URL = 'ws://192.168.1.100:8080/ws'; // Update to your backend IP/port

const CandleItem = ({ item }) => {
  const priceColor = item.Close > item.Open ? styles.priceUp : styles.priceDown;
  return (
    <View style={styles.candleItem}>
      <Text style={styles.symbol}>{item.Symbol}</Text>
      <Text style={[styles.price, priceColor]}>Price: {item.Close}</Text>
      <Text style={styles.open}>Open: {item.Open}</Text>
      <Text style={styles.high}>High: {item.High}</Text>
      <Text style={styles.low}>Low: {item.Low}</Text>
      <Text style={styles.volume}>Volume: {item.Volume}</Text>
      <Text style={styles.time}>
        Time:{' '}
        {item.Timestamps ? new Date(item.Timestamps).toLocaleString() : ''}
      </Text>
    </View>
  );
};

export default function App() {
  const [candles, setCandles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const client = new W3CWebSocket(WS_URL);
    client.onopen = () => {
      setConnected(true);
      setLoading(false);
    };
    client.onclose = () => {
      setConnected(false);
    };
    client.onmessage = message => {
      try {
        const data = JSON.parse(message.data);
        if (data.Candle) {
          setCandles(prev => {
            const filtered = prev.filter(c => c.Symbol !== data.Candle.Symbol);
            return [...filtered, data.Candle];
          });
        }
        setLoading(false);
      } catch (e) {
        // handle error
      }
    };
    return () => client.close();
  }, []);

  return (
    <SafeAreaView style={styles.container}>
      <Text style={styles.header}>Live Stock Candles</Text>
      <Text style={styles.status}>
        {connected ? 'Connected' : 'Disconnected'}
      </Text>
      {loading ? (
        <ActivityIndicator
          size="large"
          color="#007AFF"
          style={{ marginTop: 40 }}
        />
      ) : (
        <FlatList
          data={candles}
          renderItem={CandleItem}
          keyExtractor={item => item.Symbol}
          contentContainerStyle={styles.list}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
    paddingTop: 40,
  },
  header: {
    fontSize: 24,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 20,
    color: '#333',
  },
  status: {
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 10,
    color: '#007AFF',
  },
  list: {
    paddingHorizontal: 16,
  },
  candleItem: {
    backgroundColor: '#fff',
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2,
  },
  symbol: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#007AFF',
    marginBottom: 8,
  },
  price: {
    fontSize: 16,
    fontWeight: 'bold',
  },
  priceUp: {
    color: '#28a745',
  },
  priceDown: {
    color: '#d32f2f',
  },
  open: {
    fontSize: 14,
    color: '#333',
  },
  high: {
    fontSize: 14,
    color: '#007AFF',
  },
  low: {
    fontSize: 14,
    color: '#d32f2f',
  },
  volume: {
    fontSize: 16,
    color: '#666',
  },
  time: {
    fontSize: 14,
    color: '#888',
    marginTop: 4,
  },
});
