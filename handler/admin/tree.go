package admin

import "gorm.io/gorm"

// 菜单、部门都是 parent_id 自关联的树，删除与改父级共用这几个助手。

// collectDescendantIDs 迭代（BFS）收集 rootID 的全部后代 id，不含自身。
// 用 db.Model(model) 让软删除 scope 生效；visited 集合防止脏数据成环导致死循环。
func collectDescendantIDs(db *gorm.DB, model any, rootID uint) ([]uint, error) {
	var result []uint
	visited := map[uint]bool{rootID: true}
	frontier := []uint{rootID}
	for len(frontier) > 0 {
		var children []uint
		if err := db.Model(model).Where("parent_id IN ?", frontier).Pluck("id", &children).Error; err != nil {
			return nil, err
		}
		next := make([]uint, 0, len(children))
		for _, id := range children {
			if visited[id] {
				continue
			}
			visited[id] = true
			result = append(result, id)
			next = append(next, id)
		}
		frontier = next
	}
	return result, nil
}

// isSelfOrDescendant 判断把 id 的父级改成 pid 是否会成环（pid 是自身或自身的后代）。
func isSelfOrDescendant(db *gorm.DB, model any, id, pid uint) (bool, error) {
	if pid == 0 {
		return false, nil
	}
	if pid == id {
		return true, nil
	}
	descendants, err := collectDescendantIDs(db, model, id)
	if err != nil {
		return false, err
	}
	for _, d := range descendants {
		if d == pid {
			return true, nil
		}
	}
	return false, nil
}

// parentExists 校验父级存在；pid 为 0 表示根节点，恒为 true。
func parentExists(db *gorm.DB, model any, pid uint) (bool, error) {
	if pid == 0 {
		return true, nil
	}
	var count int64
	if err := db.Model(model).Where("id = ?", pid).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
