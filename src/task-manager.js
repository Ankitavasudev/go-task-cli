/**
 * Task Manager - Web Interface
 * Interactive task management with JSON export
 */

const fs = require('fs');
const path = require('path');

class Task {
    constructor(id, title, description = '', priority = 'medium', status = 'pending') {
        this.id = id;
        this.title = title;
        this.description = description;
        this.priority = priority;
        this.status = status;
        this.createdAt = new Date().toISOString();
        this.updatedAt = new Date().toISOString();
    }

    updateStatus(status) {
        this.status = status;
        this.updatedAt = new Date().toISOString();
    }

    updatePriority(priority) {
        this.priority = priority;
        this.updatedAt = new Date().toISOString();
    }

    toJSON() {
        return {
            id: this.id,
            title: this.title,
            description: this.description,
            priority: this.priority,
            status: this.status,
            createdAt: this.createdAt,
            updatedAt: this.updatedAt
        };
    }
}

class TaskManager {
    constructor(filePath) {
        this.filePath = filePath;
        this.tasks = [];
        this.loadTasks();
    }

    loadTasks() {
        try {
            if (fs.existsSync(this.filePath)) {
                const data = fs.readFileSync(this.filePath, 'utf8');
                this.tasks = JSON.parse(data);
            }
        } catch (error) {
            console.error('Error loading tasks:', error.message);
            this.tasks = [];
        }
    }

    saveTasks() {
        try {
            fs.writeFileSync(this.filePath, JSON.stringify(this.tasks, null, 2));
            return true;
        } catch (error) {
            console.error('Error saving tasks:', error.message);
            return false;
        }
    }

    addTask(title, description = '', priority = 'medium') {
        const id = this.tasks.length > 0 ? Math.max(...this.tasks.map(t => t.id)) + 1 : 1;
        const task = new Task(id, title, description, priority);
        this.tasks.push(task.toJSON());
        this.saveTasks();
        return task;
    }

    updateTask(id, updates) {
        const taskIndex = this.tasks.findIndex(t => t.id === id);
        if (taskIndex !== -1) {
            this.tasks[taskIndex] = { ...this.tasks[taskIndex], ...updates, updatedAt: new Date().toISOString() };
            this.saveTasks();
            return this.tasks[taskIndex];
        }
        return null;
    }

    deleteTask(id) {
        const taskIndex = this.tasks.findIndex(t => t.id === id);
        if (taskIndex !== -1) {
            this.tasks.splice(taskIndex, 1);
            this.saveTasks();
            return true;
        }
        return false;
    }

    getTask(id) {
        return this.tasks.find(t => t.id === id) || null;
    }

    getAllTasks() {
        return this.tasks;
    }

    getTasksByStatus(status) {
        return this.tasks.filter(t => t.status === status);
    }

    getTasksByPriority(priority) {
        return this.tasks.filter(t => t.priority === priority);
    }

    getStats() {
        const stats = {
            total: this.tasks.length,
            pending: this.tasks.filter(t => t.status === 'pending').length,
            inProgress: this.tasks.filter(t => t.status === 'in-progress').length,
            completed: this.tasks.filter(t => t.status === 'completed').length
        };
        return stats;
    }

    exportToJSON() {
        return JSON.stringify(this.tasks, null, 2);
    }
}

if (require.main === module) {
    const manager = new TaskManager('tasks.json');
    
    // Example usage
    manager.addTask('Test task', 'This is a test task', 'high');
    console.log('Tasks:', manager.getAllTasks());
    console.log('Stats:', manager.getStats());
}

module.exports = { Task, TaskManager };
