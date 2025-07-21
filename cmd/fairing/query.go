package main

// GraphQL查询或SQL查询可以定义在这里
// 根据Fairing API的实际情况调整

// 示例：如果Fairing使用GraphQL
var questionQuery = `
query GetQuestions($cursor: String, $limit: Int) {
  questions(after: $cursor, first: $limit) {
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        content
        metadata
        createdAt
        updatedAt
      }
    }
  }
}`

var responseQuery = `
query GetResponses($cursor: String, $limit: Int) {
  responses(after: $cursor, first: $limit) {
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        content
        metadata
        createdAt
        updatedAt
      }
    }
  }
}`

// TODO: 根据实际API文档调整查询结构
