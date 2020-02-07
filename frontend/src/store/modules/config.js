import config from '@/api/config';
import { MessageBox } from 'element-ui';

const state = {
  data: {}
};

// getters
const getters = {};

// actions
const actions = {
  load({commit}) {
    config.load().then(data => {
      commit('setConfig', data);
    });
  },
  save({commit}, payload) {
    config.save(payload).then(() => {
    }).catch(err => {
      MessageBox.alert(err);
    });
  }
};

// mutations
const mutations = {
  setConfig(state, config) {
    state.data = {...config};
  },
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
