const debounce = (cb) => {
  let timer;
  return function () {
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      cb();
    }, 3000);
  };
};

const processChange = debounce(() => console.log("action"));
processChange();
processChange();
processChange();
