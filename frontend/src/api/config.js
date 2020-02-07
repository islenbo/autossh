import Http from '@/http';

const http = new Http();

export default {
  load() {
    return http.get('/config');
  },
  save(config) {
    return http.post('/config', config);
  }
}
