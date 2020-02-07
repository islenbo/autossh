import axios from 'axios';

class Http {
  get(url, options = {}) {
    return this.request('get', url, options);
  }

  post(url, data, options = {}) {
    return this.request('post', url, data, options);
  }

  put(url, data, options = {}) {
    return this.request('put', url, data, options);
  }

  patch(url, data, options = {}) {
    return this.request('patch', url, data, options);
  }

  delete(url, data, options = {}) {
    return this.request('delete', url, data, options);
  }

  request(method, url, ...args) {
    return new Promise((resolve, reject) => {
      axios[method](url, ...args).then(resp => {
        switch (resp.data.code) {
          case 0:
            return resolve(resp.data.data);

          default:
            return reject(resp.data.msg);
        }
      }).catch(err => {
        return reject(err);
      });
    });
  }
}

export default Http;
