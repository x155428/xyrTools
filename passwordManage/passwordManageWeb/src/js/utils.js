export default {
     isDuplicateInSameLevel(parentNode, name) {
          // 根层
          if (!parentNode) return false;

          if (!parentNode.children || parentNode.children.length === 0) {
               return false;
          }
          return parentNode.children.some(child => child.name === name);
     }
}

