FROM node:20-alpine AS build
WORKDIR /app

# Build arguments for API URLs
ARG VITE_API_URL
ARG VITE_MSG_URL
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_MSG_URL=$VITE_MSG_URL

COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]
